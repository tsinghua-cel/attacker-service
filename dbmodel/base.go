package dbmodel

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProjectId string    `gorm:"column:project_id;index" json:"project_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	m.ProjectId = projectID
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

func ProjectFilter(db *gorm.DB) *gorm.DB {
	return db.Where("project_id = ?", projectID)
}

func ProjectFilterString() string {
	return fmt.Sprintf("project_id = '%s'", projectID)
}

func DoWithTransaction(f func(tx *gorm.DB) error) error {
	return GetDB().Transaction(f)
}
