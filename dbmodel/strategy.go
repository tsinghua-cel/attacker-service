package dbmodel

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/types"
	"gorm.io/gorm"
)

type Strategy struct {
	BaseModel
	UUID                 string  `gorm:"column:uuid;index" json:"uuid"`
	Category             string  `gorm:"column:category;size:100" json:"category"`
	Content              string  `gorm:"column:content;size:10000" json:"content"`
	MinEpoch             int64   `gorm:"column:min_epoch" json:"min_epoch"`
	MaxEpoch             int64   `gorm:"column:max_epoch" json:"max_epoch"`
	IsEnd                bool    `gorm:"column:is_end" json:"is_end"`
	ReorgCount           int     `gorm:"column:reorg_count" json:"reorg_count"`
	ImpactValidatorCount int     `gorm:"column:impact_validator_count" json:"impact_validator_count"`
	HonestLoseRateAvg    float64 `gorm:"column:honest_lose_rate_avg" json:"honest_lose_rate_avg"`
	AttackerLoseRateAvg  float64 `gorm:"column:attacker_lose_rate_avg" json:"attacker_lose_rate_avg"`
}

func (Strategy) TableName() string {
	return "t_strategy"
}

type StrategyRepository interface {
	Create(st *Strategy) error
	Update(st *Strategy) error
	GetByUUID(uuid string) *Strategy
	GetListByFilter(filters map[string]interface{}) []*Strategy
	GetSortedList(limit int, order string) []*Strategy
	GetCount() int64
}

type strategyRepositoryImpl struct {
	db *gorm.DB
}

func NewStrategyRepository(db *gorm.DB) StrategyRepository {
	return &strategyRepositoryImpl{db}
}

func (repo *strategyRepositoryImpl) Create(st *Strategy) error {
	return repo.db.Create(st).Error
}

func (repo *strategyRepositoryImpl) Update(st *Strategy) error {
	return repo.db.Save(st).Error
}

func (repo *strategyRepositoryImpl) HasByUUID(uuid string) bool {
	return len(repo.GetListByFilter(map[string]interface{}{"uuid": uuid})) > 0
}

func (repo *strategyRepositoryImpl) GetByUUID(uuid string) *Strategy {
	list := repo.GetListByFilter(map[string]interface{}{"uuid": uuid})
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func (repo *strategyRepositoryImpl) GetCount() int64 {
	var count int64
	query := ProjectFilter(repo.db).Model(&Strategy{})
	query.Where("is_end = ?", true).Count(&count)
	return count
}

func (repo *strategyRepositoryImpl) GetSortedList(limit int, order string) []*Strategy {
	list := make([]*Strategy, 0)
	query := ProjectFilter(repo.db)
	err := query.Order(order).Limit(limit).Find(&list).Error
	if err != nil {
		log.WithError(err).Error("failed to get sorted strategy list")
		return nil
	}
	return list
}

func (repo *strategyRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*Strategy {
	list := make([]*Strategy, 0)
	query := ProjectFilter(repo.db)

	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}

	query.Order("id DESC").Find(&list)
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
	if err := NewStrategyRepository(GetDB()).Create(data); err != nil {
		log.WithError(err).Error("failed to insert new strategy")
	}
}

func GetStrategyByUUID(uuid string) *Strategy {
	return NewStrategyRepository(GetDB()).GetByUUID(uuid)
}

func StrategyUpdate(st *Strategy) {
	NewStrategyRepository(GetDB()).Update(st)
}

func GetStrategyCount() int64 {
	return NewStrategyRepository(GetDB()).GetCount()
}

func GetStrategyListByReorgCount(limit int) []*Strategy {
	return NewStrategyRepository(GetDB()).GetSortedList(limit, "reorg_count DESC")
}

func GetStrategyListByHonestLoseRateAvg(limit int) []*Strategy {
	return NewStrategyRepository(GetDB()).GetSortedList(limit, "honest_lose_rate_avg DESC")
}

func GetStrategyListByGreatLostRatio(limit int) []*Strategy {
	list := make([]*Strategy, 0)
	sql := fmt.Sprintf("SELECT * FROM t_strategy WHERE attacker_lose_rate_avg != 0 AND %s ORDER BY (honest_lose_rate_avg / attacker_lose_rate_avg) DESC LIMIT %d", ProjectFilterString(), limit)
	err := GetDB().Raw(sql).Scan(&list).Error
	if err != nil {
		log.WithError(err).Error("failed to get GetStrategyListByGreatLostRatio")
		return nil
	}
	return list
}

func GetStrategyByProjectAndEpoch(project string, epoch int64) *Strategy {
	list := make([]*Strategy, 0)
	sql := fmt.Sprintf("SELECT * FROM t_strategy WHERE project_id='%s' AND min_epoch=%d", project, epoch)
	err := GetDB().Raw(sql).Scan(&list).Error
	if err != nil {
		log.WithError(err).Error("failed to get GetStrategyByProjectAndEpoch")
		return nil
	}
	if len(list) > 0 {
		return list[0]
	}
	return nil
}
