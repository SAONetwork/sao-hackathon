package model

import (
	"sao-datastore-storage/common"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

type SaoModel struct {
	Id        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Model struct {
	DB     *gorm.DB
	Config *common.Config
}

func NewModel(connString string, debug bool, config *common.Config) (*Model, error) {
	var loglevel logger.LogLevel
	if debug {
		loglevel = logger.Info
	} else {
		loglevel = logger.Silent
	}
	db, err := gorm.Open(mysql.Open(connString), &gorm.Config{Logger: logger.Default.LogMode(loglevel)})
	if err != nil {
		return nil, err
	}
	return &Model{
		DB:     db,
		Config: config,
	}, nil
}
