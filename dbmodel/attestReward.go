package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
)

type AttestReward struct {
	BaseModel
	Epoch          int64 `orm:"column(epoch)" db:"epoch" json:"epoch" form:"epoch"`                                         // epoch
	ValidatorIndex int   `orm:"column(validator_index)" db:"validator_index" json:"validator_index" form:"validator_index"` // validator index
	HeadAmount     int64 `orm:"column(head_amount)" db:"head_amount" json:"head_amount" form:"head_amount"`                 // Head reward amount
	TargetAmount   int64 `orm:"column(target_amount)" db:"target_amount" json:"target_amount" form:"target_amount"`         // Target reward amount
	SourceAmount   int64 `orm:"column(source_amount)" db:"source_amount" json:"source_amount" form:"source_amount"`         // Source reward amount.
	//Head	Target	Source	Inclusion Delay	Inactivity
}

func (AttestReward) TableName() string {
	return "t_attest_reward"
}

type AttestRewardRepository interface {
	Create(reward *AttestReward) error
	GetListByFilter(filters ...interface{}) []*AttestReward
}

type attestRewardRepositoryImpl struct {
	o orm.Ormer
}

func NewAttestRewardRepository(o orm.Ormer) AttestRewardRepository {
	return &attestRewardRepositoryImpl{o}
}

func (repo *attestRewardRepositoryImpl) Create(reward *AttestReward) error {
	reward.BeforeInsert()
	_, err := repo.o.Insert(reward)
	return err
}

func (repo *attestRewardRepositoryImpl) GetListByFilter(filters ...interface{}) []*AttestReward {
	list := make([]*AttestReward, 0)
	query := repo.o.QueryTable(new(AttestReward).TableName())
	query = ProjectFilter(query)
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	query.OrderBy("-epoch").All(&list)
	return list
}

func GetRewardListByEpoch(epoch int64) []*AttestReward {
	filters := make([]interface{}, 0)
	filters = append(filters, "epoch", epoch)
	return NewAttestRewardRepository(GetOrmInstance()).GetListByFilter(filters...)
}

func GetRewardListByValidatorIndex(index int) []*AttestReward {
	filters := make([]interface{}, 0)
	filters = append(filters, "validator_index", index)
	return NewAttestRewardRepository(GetOrmInstance()).GetListByFilter(filters...)
}

func GetRewardByValidatorAndEpoch(epoch int64, index int) *AttestReward {
	filters := make([]interface{}, 0)
	filters = append(filters, "epoch", epoch)
	filters = append(filters, "validator_index", index)

	list := NewAttestRewardRepository(GetOrmInstance()).GetListByFilter(filters...)
	if len(list) >= 0 {
		return list[0]
	}
	return nil
}

func GetMaxAttestRewardEpoch(o orm.Ormer) int64 {
	var reward AttestReward
	if o == nil {
		o = GetOrmInstance()
	}
	query := o.QueryTable(new(AttestReward).TableName())
	query = ProjectFilter(query)
	err := query.OrderBy("-epoch").One(&reward)
	if err != nil {
		return -1
	}
	return reward.Epoch
}

//func GetMaxAttestRewardedEpoch() int64 {
//	var maxEpoch int64
//	sql := fmt.Sprintf("select max(epoch) as max_epoch from %s where %s ", new(AttestReward).TableName(), ProjectFilterString())
//	if err := GetOrmInstance().Raw(sql).QueryRow(&maxEpoch); err == orm.ErrNoRows {
//		return -1
//	}
//	return maxEpoch
//}

func GetImpactValidatorCount(maxHackValIdx int, normalTargetAmount int64, epoch int64) int {
	// impact normal validator count
	var countNormal int
	sql := fmt.Sprintf("select count(1) from %s where epoch = ? and target_amount < ? and validator_index > ? and %s ", new(AttestReward).TableName(), ProjectFilterString())
	GetOrmInstance().Raw(sql, epoch, normalTargetAmount, maxHackValIdx).QueryRow(&countNormal)

	var countHacked int
	sql = fmt.Sprintf("select count(1) from %s where epoch = ? and target_amount >= ? and validator_index <= ? and %s ", new(AttestReward).TableName(), ProjectFilterString())
	GetOrmInstance().Raw(sql, epoch, normalTargetAmount, maxHackValIdx).QueryRow(&countHacked)
	return countNormal + countHacked
}

func InsertAttestRewardList(o orm.Ormer, rewards []*AttestReward) error {
	var err = DoWithTransaction(o, func(o orm.Ormer) error {
		repo := NewAttestRewardRepository(o)
		for _, reward := range rewards {
			if err := repo.Create(reward); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
