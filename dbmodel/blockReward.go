package dbmodel

import (
	"github.com/tsinghua-cel/attacker-service/common"
	"gorm.io/gorm"
)

type BlockReward struct {
	BaseModel
	Slot                   int64 `gorm:"column:slot;index" json:"slot"`
	ProposerIndex          int   `gorm:"column:proposer_index" json:"proposer_index"`
	TotalAmount            int64 `gorm:"column:total_amount" json:"total_amount"`
	AttestationAmount      int64 `gorm:"column:attestation_amount" json:"attestation_amount"`
	SyncAggregateAmount    int64 `gorm:"column:sync_aggregate_amount" json:"sync_aggregate_amount"`
	ProposerSlashingAmount int64 `gorm:"column:proposer_slashing_amount" json:"proposer_slashing_amount"`
	AttesterSlashingAmount int64 `gorm:"column:attester_slashing_amount" json:"attester_slashing_amount"`
}

func (BlockReward) TableName() string {
	return "t_block_reward"
}

type BlockRewardRepository interface {
	Create(reward *BlockReward) error
	GetListByFilter(filters map[string]interface{}) []*BlockReward
	GetListBySlotRange(start int64, end int64) []*BlockReward
}

type blockRewardRepositoryImpl struct {
	db *gorm.DB
}

func NewBlockRewardRepository(db *gorm.DB) BlockRewardRepository {
	return &blockRewardRepositoryImpl{db}
}

func (repo *blockRewardRepositoryImpl) Create(reward *BlockReward) error {
	return repo.db.Create(reward).Error
}

func (repo *blockRewardRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*BlockReward {
	list := make([]*BlockReward, 0)
	query := ProjectFilter(repo.db)

	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}

	query.Order("slot DESC").Find(&list)
	return list
}

func (repo *blockRewardRepositoryImpl) GetListBySlotRange(start int64, end int64) []*BlockReward {
	list := make([]*BlockReward, 0)
	query := ProjectFilter(repo.db)
	query.Where("slot >= ? AND slot <= ?", start, end).Order("slot ASC").Find(&list)
	return list
}

func GetMaxBlockRewardSlot(db *gorm.DB) int64 {
	var reward BlockReward
	if db == nil {
		db = GetDB()
	}

	query := ProjectFilter(db)
	err := query.Order("slot DESC").First(&reward).Error
	if err != nil {
		return -1
	}
	return reward.Slot
}

func GetBlockRewardListByEpoch(epoch int64) []*BlockReward {
	start := common.EpochStart(epoch)
	end := common.EpochEnd(epoch + 1)
	return NewBlockRewardRepository(GetDB()).GetListBySlotRange(start, end-1)
}

func InsertBlockRewardList(db *gorm.DB, rewards []*BlockReward) error {
	return DoWithTransaction(func(tx *gorm.DB) error {
		repo := NewBlockRewardRepository(tx)
		for _, reward := range rewards {
			if err := repo.Create(reward); err != nil {
				return err
			}
		}
		return nil
	})
}
