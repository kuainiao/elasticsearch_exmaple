package model

import (
	"time"
)

type DataCompany struct {
	Id             int       `xorm:"INT(11)" json:",omitempty"`
	BusinessesId   int       `xorm:"INT(11)" json:",omitempty"`
	SuppliersId    int       `xorm:"INT(11)" json:",omitempty"`
	CompanyName    string    `xorm:"TINYTEXT" json:"companyName,omitempty"`
	Country        string    `xorm:"TINYTEXT" json:"country,omitempty"`
	LinkPhone      string    `xorm:"TINYTEXT" json:"linkPhone,omitempty"`
	City           string    `xorm:"TINYTEXT" json:"city,omitempty"`
	Name           string    `xorm:"TINYTEXT" json:"name,omitempty"`
	Guid           string    `xorm:"TINYTEXT" json:"guid,omitempty"`
	ActiveContacts string    `xorm:"TINYTEXT" json:"activeContacts,omitempty"`
	CompanyWebsite string    `xorm:"TINYTEXT" json:"companyWebsite,omitempty"`
	CompanyAddress string    `xorm:"VARCHAR(500)" json:",omitempty"`
	Cantacturl     string    `xorm:"VARCHAR(200)" json:",omitempty"`
	CreateTime     time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP" json:",omitempty"`
	UpdateTime     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP" json:",omitempty"`
	IsDeleted      int       `xorm:"INT(11)" json:",omitempty"`
}
