package model

import (
	"time"
)

type ProductNew struct {
	Id                 int64     `xorm:"pk autoincr BIGINT(20)"`
	Name               string    `xorm:"not null default '' VARCHAR(50)"`
	HsCode             string    `xorm:"not null default '' index VARCHAR(20)"`
	CategoryLevel1Id   int64     `xorm:"not null default 0 BIGINT(20)"`
	CategoryLevel2Id   int64     `xorm:"not null default 0 BIGINT(20)"`
	CategoryLevel3Id   int64     `xorm:"not null default 0 BIGINT(20)"`
	CategoryLevel1Name string    `xorm:"not null default '' VARCHAR(50)"`
	CategoryLevel2Name string    `xorm:"not null default '' VARCHAR(50)"`
	CategoryLevel3Name string    `xorm:"not null default '' VARCHAR(50)"`
	UpdateTime         time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted          int       `xorm:"not null default 0 INT(11)"`
}

