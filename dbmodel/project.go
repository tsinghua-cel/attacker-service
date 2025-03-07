package dbmodel

import (
	"errors"
	"github.com/astaxie/beego/orm"
)

var (
	latestSlot int64 = -1
)

type Project struct {
	BaseModel
	StrategyCategory string `orm:"column(strategy_category)" db:"strategy_category" json:"strategy_category" form:"strategy_category"` // strategy category
	StrategyCount    int    `orm:"column(strategy_count)" db:"strategy_count" json:"strategy_count" form:"strategy_count"`             // strategy count
	LatestSlot       int64  `orm:"column(latest_slot)" db:"latest_slot" json:"latest_slot" form:"latest_slot"`                         // latest slot
}

func (Project) TableName() string {
	return "project"
}

type ProjectRepository interface {
	Create(project *Project) error
	Update(project *Project) error
	GetListByFilter(filters ...interface{}) []*Project
}

type projectRepositoryImpl struct {
	o orm.Ormer
}

func NewProjectRepository(o orm.Ormer) ProjectRepository {
	return &projectRepositoryImpl{o}
}

func (repo *projectRepositoryImpl) Create(project *Project) error {
	project.BeforeInsert()
	_, err := repo.o.Insert(project)
	return err
}

func (repo *projectRepositoryImpl) Update(project *Project) error {
	project.BeforeUpdate()
	_, err := repo.o.Update(project)
	return err
}

func (repo *projectRepositoryImpl) GetListByFilter(filters ...interface{}) []*Project {
	list := make([]*Project, 0)
	query := repo.o.QueryTable(new(Project).TableName())
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	// order by time
	query.OrderBy("-created_at").All(&list)
	return list
}

func GetProjectList() []*Project {
	return NewProjectRepository(GetOrmInstance()).GetListByFilter()
}

func NewProject() error {
	project := &Project{
		BaseModel:     BaseModel{},
		StrategyCount: 0,
	}
	return NewProjectRepository(GetOrmInstance()).Create(project)
}

func UpdateProject(project *Project) error {
	return NewProjectRepository(GetOrmInstance()).Update(project)
}

func AddStrategyCount(strategyCount int) error {
	p, err := GetProjectById(projectID)
	if err != nil {
		return err
	}

	p.StrategyCount += strategyCount

	return UpdateProject(p)
}

func SetProjectStrategyCategory(strategyCategory string) error {
	p, err := GetProjectById(projectID)
	if err != nil {
		return err
	}

	p.StrategyCategory = strategyCategory

	return UpdateProject(p)
}

func GetProjectById(id string) (*Project, error) {
	list := NewProjectRepository(GetOrmInstance()).GetListByFilter("project_id", id)
	if len(list) == 0 {
		return nil, errors.New("project not found")
	}

	return list[0], nil
}

func UpdateProjectLatestSlot(slot int64) error {
	if slot <= latestSlot {
		return nil
	}
	p, err := GetProjectById(projectID)
	if err != nil {
		return err
	}

	p.LatestSlot = slot
	return UpdateProject(p)
}
