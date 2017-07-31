package model

import (
	"fmt"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/util"
	"time"
)

/**
*
* @author willian
* @created 2017-07-27 18:13
* @email 18702515157@163.com
**/

type JsonDate time.Time
type JsonTime time.Time

func (p *JsonDate) UnmarshalJSON(data []byte) error {
	local, err := time.ParseInLocation(`"2006-01-02"`, string(data), time.Local)
	*p = JsonDate(local)
	return err
}

func (p *JsonTime) UnmarshalJSON(data []byte) error {
	local, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, string(data), time.Local)
	*p = JsonTime(local)
	return err
}

func (c JsonDate) MarshalJSON() ([]byte, error) {
	data := make([]byte, 0)
	data = append(data, '"')
	data = time.Time(c).AppendFormat(data, "2006-01-02")
	data = append(data, '"')
	return data, nil
}

func (c JsonTime) MarshalJSON() ([]byte, error) {
	data := make([]byte, 0)
	data = append(data, '"')
	data = time.Time(c).AppendFormat(data, "2006-01-02 15:04:05")
	data = append(data, '"')
	return data, nil
}

func (c JsonDate) String() string {
	return time.Time(c).Format("2006-01-02")
}

func (c JsonTime) String() string {
	return time.Time(c).Format("2006-01-02 15:04:05")
}

func Todate(in string) (out time.Time, err error) {
	out, err = time.Parse("2006-01-02", in)
	return out, err
}

func Todatetime(in string) (out time.Time, err error) {
	out, err = time.Parse("2006-01-02 15:04:05", in)
	return out, err
}

type Frank struct {
	OrderId          int64    `json:"OrderId"`
	MudiPort         string   `json:"MudiPort"`
	Purchaser        string   `json:"Purchaser"`
	PurchaserAddress string   `json:"PurchaserAddress"`
	SupplierId       int64    `json:"SupplierId"`
	PurchaserId      int64    `json:"PurchaserId"`
	OrderNo          string   `json:"OrderNo"`
	OrderWeight      float64  `json:"OrderWeight"`
	Supplier         string   `json:"Supplier"`
	ProDesc          string   `json:"ProDesc"`
	OriginalCountry  string   `json:"OriginalCountry"`
	SupplierAddress  string   `json:"SupplierAddress"`
	FrankTime        JsonDate `json:"FrankTime"`
	OrderVolume      float64  `json:"OrderVolume"`
	QiyunPort        string   `json:"QiyunPort"`
	TradeNumber      int64    `json:"TradeNumber"`
}

type Response struct {
	Error string  `json:"error"`
	Code  int     `json:"code"`
	List  []Frank `json:"list"`
	Total int64   `json:"total"`
}
