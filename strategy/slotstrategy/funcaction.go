package slotstrategy

import (
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	attaggregation "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/attestation/aggregation/attestations"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/common"
	"github.com/tsinghua-cel/attacker-service/plugins"
	"github.com/tsinghua-cel/attacker-service/types"
	"time"
)

type ActionDo func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse

type FunctionAction struct {
	doFunc ActionDo
}

func (f FunctionAction) RunAction(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
	if f.doFunc != nil {
		return f.doFunc(backend, slot, pubkey, params)
	}
	return plugins.PluginResponse{
		Cmd: types.CMD_NULL,
	}
}

func getCmdFromName(name string) types.AttackerCommand {
	switch name {
	case "null":
		return types.CMD_NULL
	case "return":
		return types.CMD_RETURN
	case "continue":
		return types.CMD_CONTINUE
	case "abort":
		return types.CMD_ABORT
	case "skip":
		return types.CMD_SKIP
	case "exit":
		return types.CMD_EXIT
	default:
		return types.CMD_NULL
	}
}

func GetFunctionAction(backend types.ServiceBackend, name string) ActionDo {
	switch name {
	case "null", "return", "continue", "abort", "skip", "exit":
		cmd := getCmdFromName(name)
		return func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
			r := plugins.PluginResponse{
				Cmd: cmd,
			}
			if len(params) > 0 {
				r.Result = params[0]
			}
			return r
		}
	case "storeSignedAttest":
		return func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
			var attestation *ethpb.Attestation
			r := plugins.PluginResponse{
				Cmd: types.CMD_NULL,
			}

			if len(params) > 0 {
				attestation = params[0].(*ethpb.Attestation)
				backend.AddSignedAttestation(uint64(slot), pubkey, attestation)
				r.Result = attestation
			}

			return r
		}

	case "delayToEpochEnd":
		return func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
			slotsPerEpoch := backend.GetSlotsPerEpoch()
			tool := common.SlotTool{
				SlotsPerEpoch: slotsPerEpoch,
			}
			epoch := tool.SlotToEpoch(slot)
			end := tool.EpochEnd(epoch)
			seconds := backend.GetIntervalPerSlot()
			total := int64(seconds) * (end - slot)
			time.Sleep(time.Second * time.Duration(total))
			r := plugins.PluginResponse{
				Cmd: types.CMD_NULL,
			}

			if len(params) > 0 {
				r.Result = params[0]
			}
			return r
		}
	case "delayHalfEpoch":
		return func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
			slotsPerEpoch := backend.GetSlotsPerEpoch()
			seconds := backend.GetIntervalPerSlot()
			total := (seconds) * (slotsPerEpoch / 2)
			time.Sleep(time.Second * time.Duration(total))
			r := plugins.PluginResponse{
				Cmd: types.CMD_NULL,
			}

			if len(params) > 0 {
				r.Result = params[0]
			}
			return r
		}
	case "rePackAttestation":
		return func(backend types.ServiceBackend, slot int64, pubkey string, params ...interface{}) plugins.PluginResponse {
			r := plugins.PluginResponse{
				Cmd: types.CMD_NULL,
			}

			if len(params) == 0 {
				return r
			}
			block := params[0].(*ethpb.SignedBeaconBlockCapella)

			tool := common.SlotTool{
				SlotsPerEpoch: backend.SlotsPerEpoch(),
			}
			epoch := tool.SlotToEpoch(slot)
			startEpoch := tool.EpochStart(epoch)
			endEpoch := tool.EpochEnd(epoch)
			attackerAttestations := make([]*ethpb.Attestation, 0)

			for i := startEpoch; i <= endEpoch; i++ {
				allSlotAttest := backend.GetAttestSet(uint64(i))
				if allSlotAttest == nil {
					continue
				}

				for publicKey, att := range allSlotAttest.Attestations {
					log.WithField("pubkey", publicKey).Debug("add attacker attestation to block")
					attackerAttestations = append(attackerAttestations, att)
				}
			}

			allAtt := append(block.Block.Body.Attestations, attackerAttestations...)
			{
				// Remove duplicates from both aggregated/unaggregated attestations. This
				// prevents inefficient aggregates being created.
				atts, _ := types.ProposerAtts(allAtt).Dedup()
				attsByDataRoot := make(map[[32]byte][]*ethpb.Attestation, len(atts))
				for _, att := range atts {
					attDataRoot, err := att.Data.HashTreeRoot()
					if err != nil {
					}
					attsByDataRoot[attDataRoot] = append(attsByDataRoot[attDataRoot], att)
				}

				attsForInclusion := types.ProposerAtts(make([]*ethpb.Attestation, 0))
				for _, as := range attsByDataRoot {
					as, err := attaggregation.Aggregate(as)
					if err != nil {
						continue
					}
					attsForInclusion = append(attsForInclusion, as...)
				}
				deduped, _ := attsForInclusion.Dedup()
				sorted, err := deduped.SortByProfitability()
				if err != nil {
					log.WithError(err).Error("sort attestation failed")
				} else {
					atts = sorted.LimitToMaxAttestations()
				}
				allAtt = atts
			}

			block.Block.Body.Attestations = allAtt

			r.Result = block
			return r
		}
	default:
		log.WithField("name", name).Error("unknown function action name")
		return nil
	}
}
