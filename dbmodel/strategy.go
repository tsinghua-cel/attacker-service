package dbmodel

import (
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"github.com/tsinghua-cel/attacker-service/types"
)

type Strategy struct {
	ID      int64  `orm:"column(id)" db:"id" json:"id" form:"id"`                     //  任务类型id
	UUID    string `orm:"column(uuid)" db:"uuid" json:"uuid" form:"uuid"`             //  策略的唯一id
	Content string `orm:"column(content)" db:"content" json:"content" form:"content"` //  策略内容
}

func (Strategy) TableName() string {
	return "t_strategy"
}

type StrategyRepository interface {
	Create(st *Strategy) error
	GetByUUID(uuid string) *Strategy
	GetListByFilter(filters ...interface{}) []*Strategy
}

type strategyRepositoryImpl struct {
	o orm.Ormer
}

func NewStrategyRepository(o orm.Ormer) StrategyRepository {
	return &strategyRepositoryImpl{o}
}

func (repo *strategyRepositoryImpl) Create(reward *Strategy) error {
	_, err := repo.o.Insert(reward)
	return err
}

func (repo *strategyRepositoryImpl) GetByUUID(uuid string) *Strategy {
	filters := make([]interface{}, 0)
	filters = append(filters, "uuid", uuid)
	return repo.GetListByFilter(filters...)[0]
}

func (repo *strategyRepositoryImpl) GetListByFilter(filters ...interface{}) []*Strategy {
	list := make([]*Strategy, 0)
	query := repo.o.QueryTable(new(Strategy).TableName())
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	query.OrderBy("-epoch").All(&list)
	return list
}

func InsertNewStrategy(st *types.Strategy) {
	d, _ := json.Marshal(st)
	data := &Strategy{
		UUID:    st.Uid,
		Content: string(d),
	}
	NewStrategyRepository(orm.NewOrm()).Create(data)
}

func GetStrategyByUUID(uuid string) *Strategy {
	return NewStrategyRepository(orm.NewOrm()).GetByUUID(uuid)
}
