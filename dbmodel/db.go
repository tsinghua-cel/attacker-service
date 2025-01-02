package dbmodel

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func DbInit(connect string) {
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
	orm.RunSyncdb("default", true, true)
}
