package dbmodel

import (
	"errors"
	"gorm.io/gorm"
)

var (
	latestSlot int64 = -1
)

type Project struct {
	BaseModel
	StrategyCategory string `gorm:"column:strategy_category" json:"strategy_category"`
	StrategyCount    int    `gorm:"column:strategy_count" json:"strategy_count"`
	LatestSlot       int64  `gorm:"column:latest_slot" json:"latest_slot"`
}

func (Project) TableName() string {
	return "project"
}

type ProjectRepository interface {
	Create(project *Project) error
	Update(project *Project) error
	GetListByFilter(filters map[string]interface{}) []*Project
}

type projectRepositoryImpl struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepositoryImpl{db}
}

func (repo *projectRepositoryImpl) Create(project *Project) error {
	return repo.db.Create(project).Error
}

func (repo *projectRepositoryImpl) Update(project *Project) error {
	return repo.db.Save(project).Error
}

func (repo *projectRepositoryImpl) GetListByFilter(filters map[string]interface{}) []*Project {
	list := make([]*Project, 0)
	query := repo.db.Model(&Project{})
	
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}
	
	query.Order("created_at DESC").Find(&list)
	return list
}

func GetProjectList() []*Project {
	return NewProjectRepository(GetDB()).GetListByFilter(map[string]interface{}{})
}

func NewProject() error {
	project := &Project{
		BaseModel:     BaseModel{},
		StrategyCount: 0,
	}
	return NewProjectRepository(GetDB()).Create(project)
}

func UpdateProject(project *Project) error {
	return NewProjectRepository(GetDB()).Update(project)
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
	list := NewProjectRepository(GetDB()).GetListByFilter(map[string]interface{}{"project_id": id})
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
