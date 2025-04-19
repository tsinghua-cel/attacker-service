package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	projectID   string
	ormInstance orm.Ormer
	once        sync.Once
)

func GetOrmInstance() orm.Ormer {
	once.Do(func() {
		ormInstance = orm.NewOrm()
	})
	return ormInstance
}

func DbInit(connect string, project_id string) {
	if project_id == "" {
		projectID = uuid.NewString()
	} else {
		projectID = project_id
	}

	// Set up database
	datasource := fmt.Sprintf("%s?charset=utf8", connect)
	orm.RegisterDriver("mysql", orm.DRMySQL)
	err := orm.RegisterDataBase("default", "mysql", datasource)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	// Configure connection pool
	orm.SetMaxIdleConns("default", 10)
	orm.SetMaxOpenConns("default", 100)

	orm.RegisterModel(new(AttestReward))
	orm.RegisterModel(new(ChainReorg))
	orm.RegisterModel(new(BlockReward))
	orm.RegisterModel(new(Strategy))
	orm.RegisterModel(new(Project))
	orm.RegisterModel(new(AttestDuty))
	orm.RegisterModel(new(BlockDuty))
	orm.RunSyncdb("default", false, false)

	// Create project
	if err = NewProject(); err != nil {
		log.WithError(err).Fatal("failed to create project")
	} else {
		log.WithField("id", projectID).Info("new project created")
	}
}
