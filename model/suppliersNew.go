package model

import (
	"time"
)

type SuppliersNew struct {
	Id              int64     `xorm:"pk autoincr BIGINT(20)" json:",omitempty"`
	Name            string    `xorm:"VARCHAR(500)" json:",omitempty"`
	OriginName      string    `xorm:"VARCHAR(500)" json:",omitempty"`
	LinkPhone       string    `xorm:"default '' VARCHAR(30)" json:",omitempty"`
	DetailsAddress  string    `xorm:"default '' VARCHAR(500)" json:",omitempty"`
	LogoUrl         string    `xorm:"default '' VARCHAR(200)" json:",omitempty"`
	Url             string    `xorm:"default '' VARCHAR(200)" json:",omitempty"`
	TotalCount      int       `xorm:"default 0 INT(11)" json:",omitempty"`
	TotalWeight     float64   `xorm:"default 0.000 DOUBLE(20,3)" json:",omitempty"`
	TotalVolume     float64   `xorm:"default 0.000 DOUBLE(20,3)" json:",omitempty"`
	LastDate        time.Time `xorm:"default '0000-00-00 00:00:00' TIMESTAMP" json:",omitempty"`
	LastProductDesc string    `xorm:"LONGTEXT" json:",omitempty"`
	LastOrderId     int64     `xorm:"default 0 BIGINT(20)" json:",omitempty"`
	DidLevel1       int64     `xorm:"not null default 0 index BIGINT(20)" json:",omitempty"`
	DidLevel2       int64     `xorm:"not null default 0 index BIGINT(20)" json:",omitempty"`
	DidLevel3       int64     `xorm:"not null default 0 index BIGINT(20)" json:",omitempty"`
	DataFlag        int       `xorm:"not null default 0 TINYINT(4)" json:",omitempty"`
	DataId          string    `xorm:"not null default '' VARCHAR(50)" json:",omitempty"`
	DataName        string    `xorm:"not null default '' VARCHAR(200)" json:",omitempty"`
	CreateTime      time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP" json:",omitempty"`
	UpdateTime      time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP" json:",omitempty"`
	IsDeleted       int       `xorm:"default 0 INT(11)" json:",omitempty"`
	DnameLevel1     string    `xorm:"default '' VARCHAR(200)" json:",omitempty"`
	DnameLevel2     string    `xorm:"default '' VARCHAR(200)" json:",omitempty"`
	DnameLevel3     string    `xorm:"default '' VARCHAR(200)" json:",omitempty"`
}
