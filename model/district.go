package model

import (
	"time"
)

type District struct {
	Did        int64     `xorm:"not null pk autoincr BIGINT(20)"`
	Dname      string    `xorm:"not null default '' index VARCHAR(50)"`
	DnameEn    string    `xorm:"not null default '' VARCHAR(50)"`
	DnameCode  string    `xorm:"default '' VARCHAR(10)"`
	Level      int       `xorm:"not null default 1 TINYINT(4)"`
	Pid        int       `xorm:"not null default 0 INT(10)"`
	Longitude  string    `xorm:"default 0.0000000 DECIMAL(10,7)"`
	Latitude   string    `xorm:"default 0.0000000 DECIMAL(10,7)"`
	CreateTime time.Time `xorm:"not null default '0000-00-00 00:00:00' DATETIME"`
	UpdateTime time.Time `xorm:"DATETIME"`
	IsDeleted  int       `xorm:"default 0 TINYINT(4)"`
}
