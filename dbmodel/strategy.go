package dbmodel

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
)

type Strategy struct {
	BaseModel
	UUID                 string  `orm:"column(uuid)" db:"uuid" json:"uuid" form:"uuid"`
	Category             string  `orm:"column(category);size(100)" db:"category" json:"category" form:"category"`
	Content              string  `orm:"column(content);size(3000)" db:"content" json:"content" form:"content"`
	MinEpoch             int64   `orm:"column(min_epoch)" db:"min_epoch" json:"min_epoch" form:"min_epoch"`
	MaxEpoch             int64   `orm:"column(max_epoch)" db:"max_epoch" json:"max_epoch" form:"max_epoch"`
	IsEnd                bool    `orm:"column(is_end)" db:"is_end" json:"is_end" form:"is_end"`
	ReorgCount           int     `orm:"column(reorg_count)" db:"reorg_count" json:"reorg_count" form:"reorg_count"`
	ImpactValidatorCount int     `orm:"column(impact_validator_count)" db:"impact_validator_count" json:"impact_validator_count" form:"impact_validator_count"`
	HonestLoseRateAvg    float64 `orm:"column(honest_lose_rate_avg)" db:"honest_lose_rate_avg" json:"honest_lose_rate_avg" form:"honest_lose_rate_avg"`
	AttackerLoseRateAvg  float64 `orm:"column(attacker_lose_rate_avg)" db:"attacker_lose_rate_avg" json:"attacker_lose_rate_avg" form:"attacker_lose_rate_avg"`
}

func (Strategy) TableName() string {
	return "t_strategy"
}

type StrategyRepository interface {
	Create(st *Strategy) error
	Update(st *Strategy) error
	GetByUUID(uuid string) *Strategy
	GetListByFilter(filters ...interface{}) []*Strategy
	GetSortedList(limit int, order string) []*Strategy
	GetCount() int64
}

type strategyRepositoryImpl struct {
	o orm.Ormer
}

func NewStrategyRepository(o orm.Ormer) StrategyRepository {
	return &strategyRepositoryImpl{o}
}

func (repo *strategyRepositoryImpl) Create(st *Strategy) error {
	st.BeforeInsert()
	_, err := repo.o.Insert(st)
	return err
}

func (repo *strategyRepositoryImpl) Update(st *Strategy) error {
	st.BeforeUpdate()
	_, err := repo.o.Update(st)
	return err
}

func (repo *strategyRepositoryImpl) HasByUUID(uuid string) bool {
	filters := make([]interface{}, 0)
	filters = append(filters, "uuid", uuid)
	return len(repo.GetListByFilter(filters...)) > 0
}

func (repo *strategyRepositoryImpl) GetByUUID(uuid string) *Strategy {
	filters := make([]interface{}, 0)
	filters = append(filters, "uuid", uuid)
	list := repo.GetListByFilter(filters...)
	if len(list) > 0 {
		return list[0]
	} else {
		return nil
	}
}

func (repo *strategyRepositoryImpl) GetCount() int64 {
	query := repo.o.QueryTable(new(Strategy).TableName())
	query = ProjectFilter(query)
	count, err := query.Filter("is_end", true).Count()
	if err != nil {
		log.WithError(err).Error("failed to get finished strategy count")
		return 0
	}
	return count
}

func (repo *strategyRepositoryImpl) GetSortedList(limit int, order string) []*Strategy {
	list := make([]*Strategy, 0)
	query := repo.o.QueryTable(new(Strategy).TableName())
	query = ProjectFilter(query)
	_, err := query.OrderBy(order).Limit(limit).All(&list)
	if err != nil {
		log.WithError(err).Error("failed to get sorted strategy list")
		return nil
	}
	return list
}

func (repo *strategyRepositoryImpl) GetListByFilter(filters ...interface{}) []*Strategy {
	list := make([]*Strategy, 0)
	query := repo.o.QueryTable(new(Strategy).TableName())
	query = ProjectFilter(query)
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	query.OrderBy("-id").All(&list)
	return list
}

func InsertNewStrategy(st *types.Strategy) {
	d, _ := json.Marshal(st)
	data := &Strategy{
		UUID:                 st.Uid,
		Content:              string(d),
		IsEnd:                false,
		ReorgCount:           0,
		ImpactValidatorCount: 0,
		Category:             st.Category,
	}
	if err := NewStrategyRepository(GetOrmInstance()).Create(data); err != nil {
		log.WithError(err).Error("failed to insert new strategy")
	}
}

func GetStrategyByUUID(uuid string) *Strategy {
	return NewStrategyRepository(GetOrmInstance()).GetByUUID(uuid)
}

func StrategyUpdate(st *Strategy) {
	NewStrategyRepository(GetOrmInstance()).Update(st)
}

func GetStrategyCount() int64 {
	return NewStrategyRepository(GetOrmInstance()).GetCount()
}

func GetStrategyListByReorgCount(limit int) []*Strategy {
	// get strategy list by reorg count desc.
	return NewStrategyRepository(GetOrmInstance()).GetSortedList(limit, "-reorg_count")
}

func GetStrategyListByHonestLoseRateAvg(limit int) []*Strategy {
	// get strategy list by honest lose rate avg desc.
	return NewStrategyRepository(GetOrmInstance()).GetSortedList(limit, "-honest_lose_rate_avg")
}

func GetStrategyListByGreatLostRatio(limit int) []*Strategy {
	// get strategy list by great honest lose rate avg desc.
	// get strategy list order by honest_lost_rate_avg/attacker_lost_rate_avg
	norm := GetOrmInstance()
	list := make([]*Strategy, 0)
	sql := fmt.Sprintf("SELECT * FROM t_strategy WHERE attacker_lose_rate_avg != 0 and %s ORDER BY (honest_lose_rate_avg / attacker_lose_rate_avg) DESC limit %d", ProjectFilterString(), limit)
	_, err := norm.Raw(sql).QueryRows(&list)
	if err != nil {
		log.WithError(err).Error("failed to get GetStrategyListByGreatLostRatio")
		return nil
	}
	return list
}

func GetStrategyByProjectAndEpoch(project string, epoch int64) *Strategy {
	norm := GetOrmInstance()
	list := make([]*Strategy, 0)
	sql := fmt.Sprintf("SELECT * FROM t_strategy WHERE project_id='%s' and min_epoch=%d", project, epoch)
	_, err := norm.Raw(sql).QueryRows(&list)
	if err != nil {
		log.WithError(err).Error("failed to get GetStrategyByProjectAndEpoch")
		return nil
	}
	if len(list) > 0 {
		return list[0]
	}
	return nil
}
