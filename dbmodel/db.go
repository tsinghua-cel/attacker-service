package dbmodel

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
)

var (
	projectID string
	db        *gorm.DB
	once      sync.Once
)

func GetDB() *gorm.DB {
	return db
}

func DbInit(connect string, project_id string) {
	if project_id == "" {
		projectID = uuid.NewString()
	} else {
		projectID = project_id
	}

	// Set up database
	datasource := fmt.Sprintf("%s?sslmode=disable", connect)
	
	var err error
	db, err = gorm.Open(postgres.Open(datasource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Fatal("failed to get database instance")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Auto migrate tables
	err = db.AutoMigrate(
		&AttestReward{},
		&ChainReorg{},
		&BlockReward{},
		&Strategy{},
		&Project{},
		&AttestDuty{},
		&BlockDuty{},
	)
	if err != nil {
		log.WithError(err).Fatal("failed to auto migrate tables")
	}

	// Create project
	if err = NewProject(); err != nil {
		log.WithError(err).Fatal("failed to create project")
	} else {
		log.WithField("id", projectID).Info("new project created")
	}
}
