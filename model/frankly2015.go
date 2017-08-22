package model

import (
	"time"
)

type Frankly2015 struct {
	Id                 int64     `xorm:"pk autoincr BIGINT(20)"`
	Gid                int64     `xorm:"default 0 BIGINT(20)"`
	FranklyCode        string    `xorm:"default '' VARCHAR(200)"`
	TotalAmount        int64     `xorm:"default 0 BIGINT(20)"`
	TotalWeight        float64   `xorm:"default 0.000 DOUBLE(15,3)"`
	WeightUnit         string    `xorm:"default 'T' VARCHAR(20)"`
	TotalVolume        string    `xorm:"default 0.000 index DECIMAL(20,3)"`
	VolumeUnit         string    `xorm:"default 'teu' VARCHAR(20)"`
	BusinessesId       int64     `xorm:"default 0 index BIGINT(20)"`
	BusinessesName     string    `xorm:"default '' VARCHAR(255)"`
	BusinessesDid      string    `xorm:"default '0' index VARCHAR(50)"`
	BusinessesAddress  string    `xorm:"default '' VARCHAR(500)"`
	SuppliersId        int64     `xorm:"default 0 index BIGINT(20)"`
	SuppliersName      string    `xorm:"default '' VARCHAR(200)"`
	SuppliersDid       string    `xorm:"default '0' index VARCHAR(50)"`
	SuppliersAddress   string    `xorm:"default '' VARCHAR(500)"`
	ProductDescription string    `xorm:"LONGTEXT"`
	FranklyTime        time.Time `xorm:"index DATETIME"`
	ContactId          int64     `xorm:"default 0 BIGINT(20)"`
	ContactName        string    `xorm:"default '' VARCHAR(50)"`
	CreateTime         time.Time `xorm:"default 'CURRENT_TIMESTAMP' DATETIME"`
	UpdateTime         time.Time `xorm:"default 'CURRENT_TIMESTAMP' DATETIME"`
	IsDeleted          int       `xorm:"default 0 INT(11)"`
	QiyunProtId        int64     `xorm:"default 0 index BIGINT(11)"`
	QiyunProtName      string    `xorm:"VARCHAR(200)"`
	MudiProtId         int64     `xorm:"default 0 BIGINT(11)"`
	MudiProtName       string    `xorm:"default '' VARCHAR(200)"`
	MudiCountry        string    `xorm:"default '' VARCHAR(20)"`
	MudiDistrictDid    int64     `xorm:"default 0 BIGINT(11)"`
	FactWeight         float64   `xorm:"default 0.000 index DOUBLE(15,3)"`
	OriginerCountry    string    `xorm:"default '' VARCHAR(200)"`
	OriginerCountryId  int64     `xorm:"default 0 BIGINT(50)"`
	VesselName         string    `xorm:"default '' VARCHAR(200)"`
	VesselId           int64     `xorm:"default 0 BIGINT(11)"`
	HsCode             string    `xorm:"default '' index VARCHAR(10)"`
}
