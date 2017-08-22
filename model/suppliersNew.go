package model

import (
	"time"
)

type SuppliersNew struct {
	Id              int64     `xorm:"pk autoincr BIGINT(20)"`
	Name            string    `xorm:"VARCHAR(500)"`
	OriginName      string    `xorm:"VARCHAR(500)"`
	LinkPhone       string    `xorm:"default '' VARCHAR(30)"`
	DetailsAddress  string    `xorm:"default '' VARCHAR(500)"`
	LogoUrl         string    `xorm:"default '' VARCHAR(200)"`
	Url             string    `xorm:"default '' VARCHAR(200)"`
	TotalCount      int       `xorm:"default 0 INT(11)"`
	TotalWeight     float64   `xorm:"default 0.000 DOUBLE(20,3)"`
	TotalVolume     float64   `xorm:"default 0.000 DOUBLE(20,3)"`
	LastDate        time.Time `xorm:"TIMESTAMP"`
	LastProductDesc string    `xorm:"LONGTEXT"`
	LastOrderId     int64     `xorm:"default 0 BIGINT(20)"`
	DidLevel1       int64     `xorm:"not null default 0 index BIGINT(20)"`
	DidLevel2       int64     `xorm:"not null default 0 index BIGINT(20)"`
	DidLevel3       int64     `xorm:"not null default 0 index BIGINT(20)"`
	DataFlag        int       `xorm:"not null default 0 TINYINT(4)"`
	DataId          string    `xorm:"not null default '' VARCHAR(50)"`
	DataName        string    `xorm:"not null default '' VARCHAR(200)"`
	CreateTime      time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	UpdateTime      time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted       int       `xorm:"default 0 INT(11)"`
	DnameLevel1     string    `xorm:"default '' VARCHAR(200)"`
	DnameLevel2     string    `xorm:"default '' VARCHAR(200)"`
	DnameLevel3     string    `xorm:"default '' VARCHAR(200)"`
}
