package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	projectID string
)

func DbInit(connect string) {
	projectID = uuid.NewString()
	// Set up database
	datasource := fmt.Sprintf("%s?charset=utf8", connect)
	orm.RegisterDriver("mysql", orm.DRMySQL)
	err := orm.RegisterDataBase("default", "mysql", datasource)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	orm.RegisterModel(new(AttestReward))
	orm.RegisterModel(new(ChainReorg))
	orm.RegisterModel(new(BlockReward))
	orm.RegisterModel(new(Strategy))
	orm.RegisterModel(new(Project))
	orm.RegisterModel(new(AttestDuty))
	orm.RegisterModel(new(BlockDuty))
	orm.RunSyncdb("default", false, true)

	// Create project
	if err = NewProject(); err != nil {
		log.WithError(err).Fatal("failed to create project")
	} else {
		log.WithField("id", projectID).Info("new project created")
	}
}
