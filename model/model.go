package model

import "time"

/**
*
* @author willian
* @created 2017-07-27 18:13
* @email 18702515157@163.com
**/

type Frank struct {
	OrderId          int64
	MudiPort         string
	Purchaser        string
	PurchaserAddress string
	SupplierId       int64
	CompanyName      string
	PurchaserId      int64
	ProductName      string
	OrderNo          string
	OrderWeight      float64
	Supplier         string
	ProDesc          string
	OriginalCountry  string
	SupplierAddress  string
	FrankTime        time.Time
	OrderVolume      float64
	QiyunPort        string
	TradeNumber      int64
}

type Response struct {
	RespMsg string
	List []Frank
	Total int64
}