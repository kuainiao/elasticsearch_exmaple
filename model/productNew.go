package model

import (
	"time"
)

type ProductNew struct {
	Id                 int       `xorm:"INT(10)"`
	Name               string    `xorm:"TINYTEXT"`
	HsCode             string    `xorm:"TINYTEXT"`
	CategoryLevel1Id   int       `xorm:"INT(10)"`
	CategoryLevel2Id   int       `xorm:"INT(10)"`
	CategoryLevel3Id   int       `xorm:"INT(10)"`
	CategoryLevel1Name string    `xorm:"TINYTEXT"`
	CategoryLevel2Name string    `xorm:"TINYTEXT"`
	CategoryLevel3Name string    `xorm:"TINYTEXT"`
	UpdateTime         time.Time `xorm:"not null default 'CURRENT_TIMESTAMP' TIMESTAMP"`
	IsDeleted          int       `xorm:"INT(11)"`
}
