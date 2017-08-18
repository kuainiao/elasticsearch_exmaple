package model

import (
	"time"
)

type CompanyDetail struct {
	Id              int       `xorm:"not null pk autoincr INT(11)"`
	CompanyType     int       `xorm:"not null INT(11)"`
	CompanyRelation string    `xorm:"not null TEXT"`
	TopTen          string    `xorm:"TEXT"`
	CompanyId       int       `xorm:"not null INT(11)"`
	ComputerTime    time.Time `xorm:"not null DATETIME"`
}
