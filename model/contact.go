package model

import (
	"time"
)

type Contact struct {
	Id           int64     `xorm:"pk autoincr BIGINT(20)" json:"id,omitempty"`
	BusinessesId int64     `xorm:"not null default 0 BIGINT(20)"`
	SuppliersId  int64     `xorm:"not null default 0 BIGINT(20)"`
	VesselId     int64     `xorm:"not null default 0 BIGINT(20)"`
	CompanyName  string    `xorm:"not null VARCHAR(200)"`
	Name         string    `xorm:"not null default '' VARCHAR(50)" json:"name,omitempty"`
	Sex          int       `xorm:"not null default 0 INT(11)" json:"sex,omitempty"`
	DeparName    string    `xorm:"not null default '' VARCHAR(100)" json:"depar_name,omitempty"`
	Position     string    `xorm:"not null default '' VARCHAR(100)" json:"position,omitempty"`
	Birthday     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	Email        string    `xorm:"not null default '' VARCHAR(50)" json:"email,omitempty"`
	Mobile       string    `xorm:"not null default '' VARCHAR(30)" json:"mobile,omitempty"`
	TelPhone     string    `xorm:"not null default '' VARCHAR(30)" json:"tel_phone,omitempty"`
	OtherLink    string    `xorm:"not null default '' VARCHAR(30)" json:"other_link,omitempty"`
	DataEmpId    string    `xorm:"not null default '0' VARCHAR(30)"`
	DataEmpGuid  string    `xorm:"not null VARCHAR(30)"`
	DataComGuid  string    `xorm:"not null default '0' index VARCHAR(30)"`
	Country      string    `xorm:"not null default '0' VARCHAR(30)" json:"country,omitempty"`
	City         string    `xorm:"not null default '0' VARCHAR(30)" json:"city,omitempty"`
	CreateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	LastTime     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	UpdateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted    int       `xorm:"not null default 0 INT(11)"`
}
