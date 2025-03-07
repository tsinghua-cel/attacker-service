package dbmodel

import (
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type BlockDuty struct {
	BaseModel
	Epoch     int64 `orm:"column(epoch)" db:"epoch" json:"epoch" form:"epoch"`
	Slot      int64 `orm:"column(slot)" db:"slot" json:"slot" form:"slot"`
	Validator int64 `orm:"column(validator)" db:"validator" json:"validator" form:"validator"`
}

func (BlockDuty) TableName() string {
	return "t_block_duty"
}

type BlockDutyRepository interface {
	Create(st *BlockDuty) error
	GetListByFilter(filters ...interface{}) []*BlockDuty
	GetSortedList(limit int, order string) []*BlockDuty
}

type blockDutyRepositoryImpl struct {
	o orm.Ormer
}

func NewBlockDutyRepository(o orm.Ormer) BlockDutyRepository {
	return &blockDutyRepositoryImpl{o}
}

func (repo *blockDutyRepositoryImpl) Create(st *BlockDuty) error {
	st.BeforeInsert()
	_, err := repo.o.Insert(st)
	return err
}

func (repo *blockDutyRepositoryImpl) GetSortedList(limit int, order string) []*BlockDuty {
	list := make([]*BlockDuty, 0)
	query := repo.o.QueryTable(new(BlockDuty).TableName())
	query = ProjectFilter(query)
	_, err := query.OrderBy(order).Limit(limit).All(&list)
	if err != nil {
		log.WithError(err).Error("failed to get block duty sorted list")
		return nil
	}
	return list
}

func (repo *blockDutyRepositoryImpl) GetListByFilter(filters ...interface{}) []*BlockDuty {
	list := make([]*BlockDuty, 0)
	query := repo.o.QueryTable(new(BlockDuty).TableName())
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

func InsertNewBlockDuties(o orm.Ormer, epoch int64, st []types.ProposerDuty) error {
	var err = DoWithTransaction(o, func(o orm.Ormer) error {
		repo := NewBlockDutyRepository(o)
		for _, s := range st {
			slot, _ := strconv.ParseInt(s.Slot, 10, 64)
			validx, _ := strconv.ParseInt(s.ValidatorIndex, 10, 64)
			data := &BlockDuty{
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

func GetBlockDuties(epoch int64) []*BlockDuty {
	o := GetOrmInstance()
	repo := NewBlockDutyRepository(o)
	return repo.GetListByFilter("epoch", epoch)
}

func GetBlockDutiesWithValidatorAndEpoch(epoch, validator int64) []*BlockDuty {
	o := GetOrmInstance()
	repo := NewBlockDutyRepository(o)
	return repo.GetListByFilter("epoch", epoch, "validator", validator)
}

func GetBlockDutiesWithValidator(validator int64) []*BlockDuty {
	o := GetOrmInstance()
	repo := NewBlockDutyRepository(o)
	return repo.GetListByFilter("validator", validator)
}

func GetMaxBlockDutyEpoch() int64 {
	o := GetOrmInstance()
	repo := NewBlockDutyRepository(o)
	list := repo.GetSortedList(1, "-epoch")
	if len(list) == 0 {
		return -1
	}
	return list[0].Epoch
}
