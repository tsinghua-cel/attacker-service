package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	log "github.com/sirupsen/logrus"
)

type DbConfig struct {
	User   string
	Passwd string
	Host   string
	Port   int
	DbName string
}

func DbInit(dbconf DbConfig) {
	// Set up database
	datasource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", dbconf.User, dbconf.Passwd, dbconf.Host, dbconf.Port, dbconf.DbName)
	orm.RegisterDriver("mysql", orm.DRMySQL)
	err := orm.RegisterDataBase("default", "mysql", datasource)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	orm.RegisterModel(new(BlockReward))
	orm.RunSyncdb("default", true, true)
}
