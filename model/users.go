package model

import (
	"time"
)

type Users struct {
	UserId          int64     `xorm:"not null pk autoincr BIGINT(20)"`
	CustId          int64     `xorm:"not null default 0 BIGINT(20)"`
	Level           int       `xorm:"not null default 0 TINYINT(4)"`
	UserName        string    `xorm:"not null default '' index(k_mp_pw) VARCHAR(80)"`
	ScreenName      string    `xorm:"not null default '' VARCHAR(80)"`
	Province        string    `xorm:"not null default '' VARCHAR(10)"`
	City            string    `xorm:"not null default '' VARCHAR(10)"`
	District        string    `xorm:"not null default '' VARCHAR(5)"`
	BirthYear       int       `xorm:"not null default 0 INT(11)"`
	BirthMonth      int       `xorm:"not null default 0 TINYINT(4)"`
	BirthDay        int       `xorm:"not null default 0 TINYINT(4)"`
	Addresss        string    `xorm:"not null default '0' VARCHAR(40)"`
	Mobilephone     int64     `xorm:"not null default 0 BIGINT(20)"`
	Telephone       int64     `xorm:"not null default 0 BIGINT(20)"`
	Longitude       string    `xorm:"not null default 0.0000000 DECIMAL(10,7)"`
	Latitude        string    `xorm:"not null default 0.0000000 DECIMAL(10,7)"`
	Gender          string    `xorm:"not null default '' VARCHAR(2)"`
	ProfileImageUrl string    `xorm:"not null default '' VARCHAR(200)"`
	OnlineStatus    int       `xorm:"not null default 0 TINYINT(4)"`
	OnlineTime      int       `xorm:"not null default 0 INT(11)"`
	Description     string    `xorm:"not null default '' VARCHAR(30)"`
	Email           string    `xorm:"not null default '' VARCHAR(30)"`
	Passwd          string    `xorm:"not null default '' index(k_mp_pw) VARCHAR(32)"`
	Company         string    `xorm:"not null default '' VARCHAR(25)"`
	Position        string    `xorm:"not null default '' VARCHAR(25)"`
	TokenContent    string    `xorm:"not null default '' VARCHAR(64)"`
	DeviceId        int       `xorm:"not null default 0 TINYINT(4)"`
	PlatformId      int       `xorm:"not null default 0 TINYINT(4)"`
	Token           string    `xorm:"not null default '' VARCHAR(40)"`
	OpenId          string    `xorm:"not null default '' VARCHAR(40)"`
	CreateTime      time.Time `xorm:"not null default '0000-00-00 00:00:00' TIMESTAMP"`
	ApplyStatus     int       `xorm:"not null default 0 TINYINT(4)"`
	UpdateTime      time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted       int       `xorm:"not null default 0 TINYINT(4)"`
}
