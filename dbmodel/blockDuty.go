package dbmodel

import (
	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type BlockDuty struct {
	BaseModel
	Epoch     int64 `gorm:"column:epoch;index" json:"epoch"`
	Slot      int64 `gorm:"column:slot;index" json:"slot"`
	Validator int64 `gorm:"column:validator;index" json:"validator"`
}

func (BlockDuty) TableName() string {
	return "t_block_duty"
}

type BlockDutyRepository interface {
	Create(st *BlockDuty) error
	GetListByFilter(filters map[string]interface{}) []*BlockDuty
	GetSortedList(limit int, order string) []*BlockDuty
}

type blockDutyRepositoryImpl struct {
	db *gorm.DB
}

func NewBlockDutyRepository(db *gorm.DB) BlockDutyRepository {
	return &blockDutyRepositoryImpl{db}
}

func (repo *blockDutyRepositoryImpl) Create(st *BlockDuty) error {
	return repo.db.Create(st).Error
}

func (repo *blockDutyRepositoryImpl) GetSortedList(limit int, order string) []*BlockDuty {
	list := make([]*BlockDuty, 0)
	query := ProjectFilter(repo.db)
	err := query.Order(order).Limit(limit).Find(&list).Error
	if err != nil {
		log.WithError(err).Error("failed to get block duty sorted list")
		return nil
	}
	return list
}

func (repo *blockDutyRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*BlockDuty {
	list := make([]*BlockDuty, 0)
	query := ProjectFilter(repo.db)
	
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}
	
	query.Order("created_at DESC").Find(&list)
	return list
}

func InsertNewBlockDuties(db *gorm.DB, epoch int64, st []types.ProposerDuty) error {
	return DoWithTransaction(func(tx *gorm.DB) error {
		repo := NewBlockDutyRepository(tx)
		for _, s := range st {
			slot, _ := strconv.ParseInt(s.Slot, 10, 64)
			validx, _ := strconv.ParseInt(s.ValidatorIndex, 10, 64)
			data := &BlockDuty{
				Slot:      slot,
				Validator: validx,
				Epoch:     epoch,
			}
			if err := repo.Create(data); err != nil {
				log.WithError(err).Error("failed to insert new block duty")
				return err
			}
		}
		return nil
	})
}

func GetBlockDuties(epoch int64) []*BlockDuty {
	repo := NewBlockDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{"epoch": epoch})
}

func GetBlockDutiesWithValidatorAndEpoch(epoch, validator int64) []*BlockDuty {
	repo := NewBlockDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{
		"epoch":     epoch,
		"validator": validator,
	})
}

func GetBlockDutiesWithValidator(validator int64) []*BlockDuty {
	repo := NewBlockDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{"validator": validator})
}

func GetMaxBlockDutyEpoch() int64 {
	repo := NewBlockDutyRepository(GetDB())
	list := repo.GetSortedList(1, "epoch DESC")
	if len(list) == 0 {
		return -1
	}
	return list[0].Epoch
}
