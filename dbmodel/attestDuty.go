package dbmodel

import (
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type AttestDuty struct {
	BaseModel
	Epoch     int64 `orm:"column(epoch)" db:"epoch" json:"epoch" form:"epoch"`
	Slot      int64 `orm:"column(slot)" db:"slot" json:"slot" form:"slot"`
	Validator int64 `orm:"column(validator)" db:"validator" json:"validator" form:"validator"`
}

func (AttestDuty) TableName() string {
	return "t_attest_duty"
}

type AttestDutyRepository interface {
	Create(st *AttestDuty) error
	GetListByFilter(filters ...interface{}) []*AttestDuty
	GetSortedList(limit int, order string) []*AttestDuty
}

type attestDutyRepositoryImpl struct {
	o orm.Ormer
}

func NewAttestDutyRepository(o orm.Ormer) AttestDutyRepository {
	return &attestDutyRepositoryImpl{o}
}

func (repo *attestDutyRepositoryImpl) Create(st *AttestDuty) error {
	st.BeforeInsert()
	_, err := repo.o.Insert(st)
	return err
}

func (repo *attestDutyRepositoryImpl) GetSortedList(limit int, order string) []*AttestDuty {
	list := make([]*AttestDuty, 0)
	query := repo.o.QueryTable(new(AttestDuty).TableName())
	query = ProjectFilter(query)
	_, err := query.OrderBy(order).Limit(limit).All(&list)
	if err != nil {
		log.WithError(err).Error("failed to get attest duty sorted list")
		return nil
	}
	return list
}

func (repo *attestDutyRepositoryImpl) GetListByFilter(filters ...interface{}) []*AttestDuty {
	list := make([]*AttestDuty, 0)
	query := repo.o.QueryTable(new(AttestDuty).TableName())
	query = ProjectFilter(query)
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	query.OrderBy("-created_at").All(&list)
	return list
}

func InsertNewAttestDuties(o orm.Ormer, epoch int64, st []types.AttestDuty) error {
	var err = DoWithTransaction(o, func(o orm.Ormer) error {
		repo := NewAttestDutyRepository(o)
		for _, s := range st {
			slot, _ := strconv.ParseInt(s.Slot, 10, 64)
			validx, _ := strconv.ParseInt(s.ValidatorIndex, 10, 64)
			data := &AttestDuty{
				Slot:      slot,
				Validator: validx,
				Epoch:     epoch,
			}
			if err := repo.Create(data); err != nil {
				log.WithError(err).Error("failed to insert new strategy")
				return err
			}
		}
		return nil
	})
	return err
}

func GetAttestDuties(epoch int64) []*AttestDuty {
	o := GetOrmInstance()
	repo := NewAttestDutyRepository(o)
	return repo.GetListByFilter("epoch", epoch)
}

func GetAttestDutiesWithValidatorAndEpoch(epoch, validator int64) []*AttestDuty {
	o := GetOrmInstance()
	repo := NewAttestDutyRepository(o)
	return repo.GetListByFilter("epoch", epoch, "validator", validator)
}

func GetAttestDutiesWithValidator(validator int64) []*AttestDuty {
	o := GetOrmInstance()
	repo := NewAttestDutyRepository(o)
	return repo.GetListByFilter("validator", validator)
}

func GetMaxAttestDutyEpoch() int64 {
	o := GetOrmInstance()
	repo := NewAttestDutyRepository(o)
	list := repo.GetSortedList(1, "-epoch")
	if len(list) == 0 {
		return -1
	}
	return list[0].Epoch
}
