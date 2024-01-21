package reward

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
	SLOTS_PER_EPOCH = "SLOTS_PER_EPOCH"
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
	if v, exist := b.config[key]; !exist {
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

func (b *BeaconGwClient) getBeaconConfig() (map[string]string, error) {
	response, err := b.doGet(fmt.Sprintf("http://%s/eth/v1/config/spec", b.endpoint))

	config := make(map[string]string)
	err = json.Unmarshal(response.Data, &b.config)
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
		b.config = config
	}
	return b.config
}

func (b *BeaconGwClient) GetLatestBeaconHeader() (BeaconResponse, error) {
	response, err := b.doGet(fmt.Sprintf("http://%s/eth/v1/beacon/headers", b.endpoint))
	return response, err
}

// default grpc-gateway port is 3500
func (b *BeaconGwClient) GetAllValReward(slot int) (BeaconResponse, error) {
	url := fmt.Sprintf("http://%s/eth/v1/beacon/rewards/attestations/%d", b.endpoint, slot)
	resp, err := httplib.Post(url).Body([]byte("[]")).Response()
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

func (b *BeaconGwClient) GetValReward(slot int, valIdxs []int) (BeaconResponse, error) {
	url := fmt.Sprintf("http://%s/eth/v1/beacon/rewards/attestations/%d", b.endpoint, slot)
	vals := make([]string, len(valIdxs))
	for i := 0; i < len(valIdxs); i++ {
		vals[i] = strconv.FormatInt(int64(valIdxs[i]), 10)
	}
	d, err := json.Marshal(vals)
	if err != nil {
		log.WithError(err).Error("get reward failed when marshal vals")
		return BeaconResponse{}, err
	}
	resp, err := httplib.Post(url).Body(d).Response()
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
