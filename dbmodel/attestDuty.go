package dbmodel

import (
	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type AttestDuty struct {
	BaseModel
	Epoch     int64 `gorm:"column:epoch;index" json:"epoch"`
	Slot      int64 `gorm:"column:slot;index" json:"slot"`
	Validator int64 `gorm:"column:validator;index" json:"validator"`
}

func (AttestDuty) TableName() string {
	return "t_attest_duty"
}

type AttestDutyRepository interface {
	Create(st *AttestDuty) error
	GetListByFilter(filters map[string]interface{}) []*AttestDuty
	GetSortedList(limit int, order string) []*AttestDuty
}

type attestDutyRepositoryImpl struct {
	db *gorm.DB
}

func NewAttestDutyRepository(db *gorm.DB) AttestDutyRepository {
	return &attestDutyRepositoryImpl{db}
}

func (repo *attestDutyRepositoryImpl) Create(st *AttestDuty) error {
	return repo.db.Create(st).Error
}

func (repo *attestDutyRepositoryImpl) GetSortedList(limit int, order string) []*AttestDuty {
	list := make([]*AttestDuty, 0)
	query := ProjectFilter(repo.db)
	err := query.Order(order).Limit(limit).Find(&list).Error
	if err != nil {
		log.WithError(err).Error("failed to get attest duty sorted list")
		return nil
	}
	return list
}

func (repo *attestDutyRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*AttestDuty {
	list := make([]*AttestDuty, 0)
	query := ProjectFilter(repo.db)
	
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}
	
	query.Order("created_at DESC").Find(&list)
	return list
}

func InsertNewAttestDuties(db *gorm.DB, epoch int64, st []types.AttestDuty) error {
	return DoWithTransaction(func(tx *gorm.DB) error {
		repo := NewAttestDutyRepository(tx)
		for _, s := range st {
			slot, _ := strconv.ParseInt(s.Slot, 10, 64)
			validx, _ := strconv.ParseInt(s.ValidatorIndex, 10, 64)
			data := &AttestDuty{
				Slot:      slot,
				Validator: validx,
				Epoch:     epoch,
			}
			if err := repo.Create(data); err != nil {
				log.WithError(err).Error("failed to insert new attest duty")
				return err
			}
		}
		return nil
	})
}

func GetAttestDuties(epoch int64) []*AttestDuty {
	repo := NewAttestDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{"epoch": epoch})
}

func GetAttestDutiesWithValidatorAndEpoch(epoch, validator int64) []*AttestDuty {
	repo := NewAttestDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{
		"epoch":     epoch,
		"validator": validator,
	})
}

func GetAttestDutiesWithValidator(validator int64) []*AttestDuty {
	repo := NewAttestDutyRepository(GetDB())
	return repo.GetListByFilter(map[string]interface{}{"validator": validator})
}

func GetMaxAttestDutyEpoch() int64 {
	repo := NewAttestDutyRepository(GetDB())
	list := repo.GetSortedList(1, "epoch DESC")
	if len(list) == 0 {
		return -1
	}
	return list[0].Epoch
}
