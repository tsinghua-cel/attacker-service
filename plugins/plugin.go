package plugins

import (
	"context"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
)

type PluginContext interface {
	Context() context.Context
	Backend() types.ServiceBackend
	Logger() logrus.Logger
}

type PluginResponse struct {
	Cmd    types.AttackerCommand
	Result interface{}
}

type AttackerPlugin interface {
	AttestBeforeBroadCast(ctx PluginContext, slot uint64) PluginResponse
	AttestAfterBroadCast(ctx PluginContext, slot uint64) PluginResponse
	AttestBeforeSign(ctx PluginContext, slot uint64, pubkey string, attestData *ethpb.AttestationData) PluginResponse
	AttestAfterSign(ctx PluginContext, slot uint64, pubkey string, attest *ethpb.Attestation) PluginResponse
	AttestBeforePropose(ctx PluginContext, slot uint64, pubkey string, attest *ethpb.Attestation) PluginResponse
	AttestAfterPropose(ctx PluginContext, slot uint64, pubkey string, attest *ethpb.Attestation) PluginResponse

	BlockBroadCastDelay(ctx PluginContext) PluginResponse
	BlockDelayForReceiveBlock(ctx PluginContext, slot uint64) PluginResponse
	BlockBeforeBroadCast(ctx PluginContext, slot uint64) PluginResponse
	BlockAfterBroadCast(ctx PluginContext, slot uint64) PluginResponse
	BlockBeforeMakeBlock(ctx PluginContext, slot uint64, pubkey string) PluginResponse
	BlockBeforeSign(ctx PluginContext, slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock_Capella) PluginResponse
	BlockAfterSign(ctx PluginContext, slot uint64, pubkey string, block *ethpb.GenericSignedBeaconBlock_Capella) PluginResponse
	BlockBeforePropose(ctx PluginContext, slot uint64, pubkey string, block *ethpb.SignedBeaconBlockCapella) PluginResponse
	BlockAfterPropose(ctx PluginContext, slot uint64, pubkey string, block *ethpb.SignedBeaconBlockCapella) PluginResponse
}
