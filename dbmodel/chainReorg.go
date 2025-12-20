package dbmodel

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/tsinghua-cel/attacker-service/types"
)

type ChainReorg struct {
	BaseModel
	Epoch                 int64  `gorm:"column:epoch;index" json:"epoch"`
	Slot                  int64  `gorm:"column:slot;index" json:"slot"`
	Depth                 int    `gorm:"column:depth" json:"depth"`
	OldBlockSlot          int64  `gorm:"column:old_block_slot" json:"old_block_slot"`
	NewBlockSlot          int64  `gorm:"column:new_block_slot" json:"new_block_slot"`
	OldBlockProposerIndex int64  `gorm:"column:old_block_proposer_index" json:"old_block_proposer_index"`
	NewBlockProposerIndex int64  `gorm:"column:new_block_proposer_index" json:"new_block_proposer_index"`
	OldHeadState          string `gorm:"column:old_head_state" json:"old_head_state"`
	NewHeadState          string `gorm:"column:new_head_state" json:"new_head_state"`
}

func (ChainReorg) TableName() string {
	return "t_chain_reorg"
}

type ChainReorgRepository interface {
	Create(reorg *ChainReorg) error
	GetListByFilter(filters map[string]interface{}) []*ChainReorg
}

type chainReorgRepositoryImpl struct {
	db *gorm.DB
}

func NewChainReorgRepository(db *gorm.DB) ChainReorgRepository {
	return &chainReorgRepositoryImpl{db}
}

func (repo *chainReorgRepositoryImpl) Create(reorg *ChainReorg) error {
	return repo.db.Create(reorg).Error
}

func (repo *chainReorgRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*ChainReorg {
	list := make([]*ChainReorg, 0)
	query := ProjectFilter(repo.db)
	
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}
	
	query.Order("slot DESC").Find(&list)
	return list
}

func InsertNewReorg(ev types.ReorgEvent) {
	NewChainReorgRepository(GetDB()).Create(&ChainReorg{
		Epoch:                 int64(ev.Epoch),
		Slot:                  int64(ev.Slot),
		Depth:                 int(ev.Depth),
		OldBlockSlot:          ev.OldBlockSlot,
		NewBlockSlot:          ev.NewBlockSlot,
		OldBlockProposerIndex: ev.OldBlockProposerIndex,
		NewBlockProposerIndex: ev.NewBlockProposerIndex,
		OldHeadState:          ev.OldHeadState,
		NewHeadState:          ev.NewHeadState,
	})
}

func GetAllReorgList() []*ChainReorg {
	return NewChainReorgRepository(GetDB()).GetListByFilter(map[string]interface{}{})
}

func GetReorgListByEpoch(epoch int64) []*ChainReorg {
	return NewChainReorgRepository(GetDB()).GetListByFilter(map[string]interface{}{"epoch": epoch})
}

func GetReorgCountByEpoch(epoch int64) int {
	var count int64
	sql := fmt.Sprintf("SELECT count(1) FROM %s WHERE epoch = ? AND %s", new(ChainReorg).TableName(), ProjectFilterString())
	GetDB().Raw(sql, epoch).Scan(&count)
	return int(count)
}
