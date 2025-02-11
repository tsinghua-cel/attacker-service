package dbmodel

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type BaseModel struct {
	ID        int64     `orm:"column(id)" db:"id" json:"id" form:"id"`                                 // uniq id
	ProjectId string    `orm:"column(project_id)" db:"project_id" json:"project_id" form:"project_id"` // project id
	CreatedAt time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"created_at"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updated_at"`
}

func (m *BaseModel) BeforeInsert() {
	m.ProjectId = projectID
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
}

func (m *BaseModel) BeforeUpdate() {
	m.UpdatedAt = time.Now()
}

func ProjectFilter(query orm.QuerySeter) orm.QuerySeter {
	return query.Filter("project_id", projectID)
}

func ProjectFilterString() string {
	return "project_id = " + projectID
}
