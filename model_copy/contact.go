package model

import (
	"time"
)

type Contact struct {
	Id           int64     `xorm:"pk autoincr BIGINT(20)" json:"id,omitempty"`
	BusinessesId int64     `xorm:"not null default 0 BIGINT(20)" json:"businesses_id,omitempty"`
	SuppliersId  int64     `xorm:"not null default 0 BIGINT(20)" json:"suppliers_id,omitempty"`
	VesselId     int64     `xorm:"not null default 0 BIGINT(20)" json:"vessel_id,omitempty"`
	CompanyName  string    `xorm:"not null VARCHAR(200)" json:"company_name,omitempty"`
	Name         string    `xorm:"not null default '' VARCHAR(50)" json:"name,omitempty"`
	Sex          int       `xorm:"not null default 0 INT(11)" json:"sex,omitempty"`
	DeparName    string    `xorm:"not null default '' VARCHAR(100)" json:"depar_name,omitempty"`
	Position     string    `xorm:"not null default '' VARCHAR(100)" json:"position,omitempty"`
	Birthday     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	Email        string    `xorm:"not null default '' VARCHAR(50)" json:"email"`
	Mobile       string    `xorm:"not null default '' VARCHAR(30)" json:"mobile,omitempty"`
	TelPhone     string    `xorm:"not null default '' VARCHAR(30)" json:"tel_phone"`
	OtherLink    string    `xorm:"not null default '' VARCHAR(30)" json:"other_link,omitempty"`
	DataEmpId    string    `xorm:"not null default '0' VARCHAR(30)" json:"data_emp_id,omitempty"`
	DataEmpGuid  string    `xorm:"not null VARCHAR(30)" json:"data_emp_guid,omitempty"`
	DataComGuid  string    `xorm:"not null default '0' index VARCHAR(30)" json:"data_com_guid,omitempty"`
	Country      string    `xorm:"not null default '0' VARCHAR(30)" json:"country,omitempty"`
	City         string    `xorm:"not null default '0' VARCHAR(30)" json:"city,omitempty"`
	CreateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"create_time,omitempty"`
	LastTime     time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP" json:"last_time,omitempty"`
	UpdateTime   time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"update_time,omitempty"`
	IsDeleted    int       `xorm:"not null default 0 INT(11)" json:"is_deleted,omitempty"`
}
