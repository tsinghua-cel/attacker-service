package dbmodel

import (
	"fmt"
	"gorm.io/gorm"
)

type AttestReward struct {
	BaseModel
	Epoch          int64 `gorm:"column:epoch;index" json:"epoch"`
	ValidatorIndex int   `gorm:"column:validator_index;index" json:"validator_index"`
	HeadAmount     int64 `gorm:"column:head_amount" json:"head_amount"`
	TargetAmount   int64 `gorm:"column:target_amount" json:"target_amount"`
	SourceAmount   int64 `gorm:"column:source_amount" json:"source_amount"`
}

func (AttestReward) TableName() string {
	return "t_attest_reward"
}

type AttestRewardRepository interface {
	Create(reward *AttestReward) error
	GetListByFilter(filters map[string]interface{}) []*AttestReward
}

type attestRewardRepositoryImpl struct {
	db *gorm.DB
}

func NewAttestRewardRepository(db *gorm.DB) AttestRewardRepository {
	return &attestRewardRepositoryImpl{db}
}

func (repo *attestRewardRepositoryImpl) Create(reward *AttestReward) error {
	return repo.db.Create(reward).Error
}

func (repo *attestRewardRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*AttestReward {
	list := make([]*AttestReward, 0)
	query := ProjectFilter(repo.db)
	
	for k, v := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", k), v)
	}
	
	query.Order("epoch DESC").Find(&list)
	return list
}

func GetRewardListByEpoch(epoch int64) []*AttestReward {
	filters := map[string]interface{}{
		"epoch": epoch,
	}
	return NewAttestRewardRepository(GetDB()).GetListByFilter(filters)
}

func GetRewardListByValidatorIndex(index int) []*AttestReward {
	filters := map[string]interface{}{
		"validator_index": index,
	}
	return NewAttestRewardRepository(GetDB()).GetListByFilter(filters)
}

func GetRewardByValidatorAndEpoch(epoch int64, index int) *AttestReward {
	filters := map[string]interface{}{
		"epoch":           epoch,
		"validator_index": index,
	}

	list := NewAttestRewardRepository(GetDB()).GetListByFilter(filters)
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func GetMaxAttestRewardEpoch(db *gorm.DB) int64 {
	var reward AttestReward
	if db == nil {
		db = GetDB()
	}
	
	query := ProjectFilter(db)
	err := query.Order("epoch DESC").First(&reward).Error
	if err != nil {
		return -1
	}
	return reward.Epoch
}

func GetImpactValidatorCount(maxHackValIdx int, normalTargetAmount int64, epoch int64) int {
	// impact normal validator count
	var countNormal int64
	sql := fmt.Sprintf("SELECT count(1) FROM %s WHERE epoch = ? AND target_amount < ? AND validator_index > ? AND %s", 
		new(AttestReward).TableName(), ProjectFilterString())
	GetDB().Raw(sql, epoch, normalTargetAmount, maxHackValIdx).Scan(&countNormal)

	var countHacked int64
	sql = fmt.Sprintf("SELECT count(1) FROM %s WHERE epoch = ? AND target_amount >= ? AND validator_index <= ? AND %s", 
		new(AttestReward).TableName(), ProjectFilterString())
	GetDB().Raw(sql, epoch, normalTargetAmount, maxHackValIdx).Scan(&countHacked)
	
	return int(countNormal + countHacked)
}

func InsertAttestRewardList(db *gorm.DB, rewards []*AttestReward) error {
	return DoWithTransaction(func(tx *gorm.DB) error {
		repo := NewAttestRewardRepository(tx)
		for _, reward := range rewards {
			if err := repo.Create(reward); err != nil {
				return err
			}
		}
		return nil
	})
}
