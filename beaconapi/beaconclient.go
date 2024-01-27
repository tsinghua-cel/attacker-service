package beaconapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcutil "github.com/prysmaticlabs/prysm/v4/api/grpc"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"strconv"
	"time"
)

const (
	SLOTS_PER_EPOCH  = "SLOTS_PER_EPOCH"
	SECONDS_PER_SLOT = "SECONDS_PER_SLOT"
)

type BeaconGwClient struct {
	endpoint string
	config   map[string]string
}

func NewBeaconGwClient(endpoint string) *BeaconGwClient {
	return &BeaconGwClient{
		endpoint: endpoint,
		config:   make(map[string]string),
	}
}

func (b *BeaconGwClient) GetIntConfig(key string) (int, error) {
	config := b.GetBeaconConfig()
	if v, exist := config[key]; !exist {
		return 0, nil
	} else {
		return strconv.Atoi(v)
	}
}

func (b *BeaconGwClient) doGet(url string) (BeaconResponse, error) {
	resp, err := httplib.Get(url).Response()
	if err != nil {
		return BeaconResponse{}, err
	}
	defer resp.Body.Close()

	var response BeaconResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.WithError(err).Error("Error decoding response")
	}
	return response, nil
}

func (b *BeaconGwClient) doPost(url string, data []byte) (BeaconResponse, error) {
	resp, err := httplib.Post(url).Body(data).Response()
	if err != nil {
		return BeaconResponse{}, err
	}
	defer resp.Body.Close()

	var response BeaconResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.WithError(err).Error("Error decoding response")
	}
	return response, nil
}

func (b *BeaconGwClient) getBeaconConfig() (map[string]interface{}, error) {
	response, err := b.doGet(fmt.Sprintf("http://%s/eth/v1/config/spec", b.endpoint))

	config := make(map[string]interface{})
	err = json.Unmarshal(response.Data, &config)
	if err != nil {
		log.WithError(err).Error("unmarshal config data failed")
	}
	return config, nil
}

func (b *BeaconGwClient) GetBeaconConfig() map[string]string {
	if len(b.config) == 0 {
		config, err := b.getBeaconConfig()
		if err != nil {
			// todo: add log
			return nil
		}
		b.config = make(map[string]string)
		for key, v := range config {
			b.config[key] = v.(string)
		}
	}
	return b.config
}

func (b *BeaconGwClient) GetLatestBeaconHeader() (BeaconHeaderInfo, error) {
	response, err := b.doGet(fmt.Sprintf("http://%s/eth/v1/beacon/headers", b.endpoint))
	var headers = make([]BeaconHeaderInfo, 0)
	err = json.Unmarshal(response.Data, &headers)
	if err != nil {
		// todo: add log.
		return BeaconHeaderInfo{}, err
	}

	return headers[0], nil
}

// default grpc-gateway port is 3500
func (b *BeaconGwClient) GetAllValReward(epoch int) ([]TotalReward, error) {
	url := fmt.Sprintf("http://%s/eth/v1/beacon/rewards/attestations/%d", b.endpoint, epoch)
	response, err := b.doPost(url, []byte("[]"))
	var rewardInfo RewardInfo
	err = json.Unmarshal(response.Data, &rewardInfo)
	if err != nil {
		log.WithError(err).Error("unmarshal reward data failed")
		return nil, err
	}
	return rewardInfo.TotalRewards, err
}

func (b *BeaconGwClient) GetValReward(epoch int, valIdxs []int) (BeaconResponse, error) {
	url := fmt.Sprintf("http://%s/eth/v1/beacon/rewards/attestations/%d", b.endpoint, epoch)
	vals := make([]string, len(valIdxs))
	for i := 0; i < len(valIdxs); i++ {
		vals[i] = strconv.FormatInt(int64(valIdxs[i]), 10)
	}
	d, err := json.Marshal(vals)
	if err != nil {
		log.WithError(err).Error("get reward failed when marshal vals")
		return BeaconResponse{}, err
	}
	response, err := b.doPost(url, d)
	return response, err
}

// /eth/v1/validator/duties/proposer/:epoch
func (b *BeaconGwClient) GetProposerDuties(epoch int) ([]ProposerDuty, error) {
	url := fmt.Sprintf("http://%s/eth/v1/validator/duties/proposer/%d", b.endpoint, epoch)
	var duties = make([]ProposerDuty, 0)

	response, err := b.doGet(url)
	err = json.Unmarshal(response.Data, &duties)
	if err != nil {
		// todo: add log.
		return []ProposerDuty{}, err
	}

	return duties, err
}

// POST /eth/v1/validator/duties/attester/:epoch
func (b *BeaconGwClient) GetAttesterDuties(epoch int, vals []int) ([]AttestDuty, error) {
	url := fmt.Sprintf("http://%s/eth/v1/validator/duties/attester/%d", b.endpoint, epoch)
	param := make([]string, len(vals))
	for i := 0; i < len(vals); i++ {
		param[i] = strconv.FormatInt(int64(vals[i]), 10)
	}
	paramData, _ := json.Marshal(param)
	var duties = make([]AttestDuty, 0)

	response, err := b.doPost(url, paramData)
	err = json.Unmarshal(response.Data, &duties)
	if err != nil {
		return []AttestDuty{}, err
	}
	return duties, err
}

func (b *BeaconGwClient) GetNextEpochProposerDuties() ([]ProposerDuty, error) {
	latestHeader, err := b.GetLatestBeaconHeader()
	if err != nil {
		return nil, err
	}
	slotPerEpoch, _ := b.GetIntConfig(SLOTS_PER_EPOCH)
	curSlot, _ := strconv.Atoi(latestHeader.Header.Message.Slot)
	epoch := curSlot / slotPerEpoch
	return b.GetProposerDuties(epoch + 1)
}

func (b *BeaconGwClient) GetCurrentEpochProposerDuties() ([]ProposerDuty, error) {
	latestHeader, err := b.GetLatestBeaconHeader()
	if err != nil {
		return nil, err
	}
	slotPerEpoch, _ := b.GetIntConfig(SLOTS_PER_EPOCH)
	curSlot, _ := strconv.Atoi(latestHeader.Header.Message.Slot)
	epoch := curSlot / slotPerEpoch
	return b.GetProposerDuties(epoch)
}

func (b *BeaconGwClient) GetCurrentEpochAttestDuties() ([]AttestDuty, error) {
	latestHeader, err := b.GetLatestBeaconHeader()
	if err != nil {
		return nil, err
	}
	slotPerEpoch, _ := b.GetIntConfig(SLOTS_PER_EPOCH)
	curSlot, _ := strconv.Atoi(latestHeader.Header.Message.Slot)
	epoch := curSlot / slotPerEpoch
	vals := make([]int, 64)
	for i := 0; i < len(vals); i++ {
		vals[i] = i
	}
	return b.GetAttesterDuties(epoch, vals)
}

func (b *BeaconGwClient) GetNextEpochAttestDuties() ([]AttestDuty, error) {
	latestHeader, err := b.GetLatestBeaconHeader()
	if err != nil {
		return nil, err
	}
	slotPerEpoch, _ := b.GetIntConfig(SLOTS_PER_EPOCH)
	curSlot, _ := strconv.Atoi(latestHeader.Header.Message.Slot)
	epoch := curSlot / slotPerEpoch
	vals := make([]int, 64)
	for i := 0; i < len(vals); i++ {
		vals[i] = i
	}
	return b.GetAttesterDuties(epoch+1, vals)
}

func (b *BeaconGwClient) GetAllValidators() {
	//vals := make([]int, 64)
	//for i := 0; i < len(vals); i++ {
	//	vals[i] = i
	//}
	//duties :=

}

// default grpc port is 4000
func GetValidators(grpcEndpoint string) ([]string, error) {
	client := NewGrpcBeaconChainClient(grpcEndpoint)
	req := &ethpb.ListValidatorsRequest{}
	vals, err := client.ListValidators(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var pubkeys = make([]string, 0, len(vals.ValidatorList))
	for _, val := range vals.ValidatorList {
		pubkeys = append(pubkeys, hex.EncodeToString(val.Validator.PublicKey))
	}
	return pubkeys, nil
}

// ConstructDialOptions constructs a list of grpc dial options
func ConstructDialOptions(
	maxCallRecvMsgSize int,
	withCert string,
	grpcRetries uint,
	grpcRetryDelay time.Duration,
	extraOpts ...grpc.DialOption,
) []grpc.DialOption {
	var transportSecurity grpc.DialOption
	if withCert != "" {
		creds, err := credentials.NewClientTLSFromFile(withCert, "")
		if err != nil {
			log.WithError(err).Error("Could not get valid credentials")
			return nil
		}
		transportSecurity = grpc.WithTransportCredentials(creds)
	} else {
		transportSecurity = grpc.WithInsecure()
		log.Warn("You are using an insecure gRPC connection. If you are running your beacon node and " +
			"validator on the same machines, you can ignore this message. If you want to know " +
			"how to enable secure connections, see: https://docs.prylabs.network/docs/prysm-usage/secure-grpc")
	}

	if maxCallRecvMsgSize == 0 {
		maxCallRecvMsgSize = 10 * 5 << 20 // Default 50Mb
	}

	dialOpts := []grpc.DialOption{
		transportSecurity,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
			grpcretry.WithMax(grpcRetries),
			grpcretry.WithBackoff(grpcretry.BackoffLinear(grpcRetryDelay)),
		),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
			grpcopentracing.UnaryClientInterceptor(),
			grpcprometheus.UnaryClientInterceptor,
			grpcretry.UnaryClientInterceptor(),
			grpcutil.LogRequests,
		)),
		grpc.WithChainStreamInterceptor(
			grpcutil.LogStream,
			grpcopentracing.StreamClientInterceptor(),
			grpcprometheus.StreamClientInterceptor,
			grpcretry.StreamClientInterceptor(),
		),
		grpc.WithResolvers(&multipleEndpointsGrpcResolverBuilder{}),
	}

	dialOpts = append(dialOpts, extraOpts...)
	return dialOpts
}

func NewGrpcBeaconChainClient(endpoint string) ethpb.BeaconChainClient {

	dialOpts := ConstructDialOptions(
		1024*1024,
		"",
		5,
		time.Second,
	)
	if dialOpts == nil {
		return nil
	}

	grpcConn, err := grpc.DialContext(context.Background(), endpoint, dialOpts...)
	if err != nil {
		return nil
	}
	return ethpb.NewBeaconChainClient(grpcConn)
}
