package model

import (
	"time"
)

type Contact struct {
	Id           int64     `xorm:"pk autoincr BIGINT(20)"`
	BusinessesId int64     `xorm:"not null default 0 BIGINT(20)"`
	SuppliersId  int64     `xorm:"not null default 0 BIGINT(20)"`
	VesselId     int64     `xorm:"not null default 0 BIGINT(20)"`
	CompanyName  string    `xorm:"not null VARCHAR(200)"`
	Name         string    `xorm:"not null default '' VARCHAR(50)"`
	Sex          int       `xorm:"not null default 0 INT(11)"`
	DeparName    string    `xorm:"not null default '' VARCHAR(100)"`
	Position     string    `xorm:"not null default '' VARCHAR(100)"`
	Birthday     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	Email        string    `xorm:"not null default '' VARCHAR(50)"`
	Mobile       string    `xorm:"not null default '' VARCHAR(30)"`
	TelPhone     string    `xorm:"not null default '' VARCHAR(30)"`
	OtherLink    string    `xorm:"not null default '' VARCHAR(30)"`
	DataEmpId    string    `xorm:"not null default '0' VARCHAR(30)"`
	DataEmpGuid  string    `xorm:"not null VARCHAR(30)"`
	DataComGuid  string    `xorm:"not null default '0' index VARCHAR(30)"`
	Country      string    `xorm:"not null default '0' VARCHAR(30)"`
	City         string    `xorm:"not null default '0' VARCHAR(30)"`
	CreateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	LastTime     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	UpdateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted    int       `xorm:"not null default 0 INT(11)"`
}
