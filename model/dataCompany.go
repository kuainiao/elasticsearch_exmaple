package model

import (
	"time"
)

type DataCompany struct {
	Id             int       `xorm:"INT(11)"`
	BusinessesId   int       `xorm:"INT(11)"`
	SuppliersId    int       `xorm:"INT(11)"`
	CompanyName    string    `xorm:"TINYTEXT"`
	Country        string    `xorm:"TINYTEXT"`
	LinkPhone      string    `xorm:"TINYTEXT"`
	City           string    `xorm:"TINYTEXT"`
	Name           string    `xorm:"TINYTEXT"`
	Guid           string    `xorm:"TINYTEXT"`
	ActiveContacts string    `xorm:"TINYTEXT"`
	CompanyWebsite string    `xorm:"TINYTEXT"`
	CompanyAddress string    `xorm:"VARCHAR(500)"`
	Cantacturl     string    `xorm:"VARCHAR(200)"`
	CreateTime     time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	UpdateTime     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	IsDeleted      int       `xorm:"INT(11)"`
}
