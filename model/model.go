package model

import (
	"time"
)

/**
*
* @author willian
* @created 2017-07-27 18:13
* @email 18702515157@163.com
**/

//xorm reverse mysql "root:19960118@tcp(127.0.0.1:3306)/web"
//$GOPATH/src/github.com/go-xorm/cmd/xorm/templates/goxorm
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
	OrderId              int64    `json:"OrderId,omitempty"`
	MudiPort             string   `json:"MudiPort,omitempty"`
	Purchaser            string   `json:"Purchaser,omitempty"`
	PurchaserAddress     string   `json:"PurchaserAddress,omitempty"`
	SupplierId           int64    `json:"SupplierId,omitempty"`
	PurchaserId          int64    `json:"PurchaserId,omitempty"`
	OrderNo              string   `json:"OrderNo,omitempty"`
	OrderWeight          float64  `json:"OrderWeight,omitempty"`
	Supplier             string   `json:"Supplier,omitempty"`
	ProDesc              string   `json:"ProDesc,omitempty"`
	ProKey               string   `json:"pro_key,omitempty"`
	OriginalCountry      string   `json:"OriginalCountry,omitempty"`
	SupplierAddress      string   `json:"SupplierAddress,omitempty"`
	FrankTime            JsonDate `json:"FrankTime,omitempty"`
	OrderVolume          float64  `json:"OrderVolume,omitempty"`
	QiyunPort            string   `json:"QiyunPort,omitempty"`
	TradeNumber          int64    `json:"TradeNumber,omitempty"`
	ProductName          string   `json:"ProductName,omitempty"`
	CompanyName          string   `json:"company_name,omitempty"`
	CompanyId            int64    `json:"company_id,omitempty"`
	Score                int      `json:"score,omitempty"`
	CompanyAddress       string   `json:"company_address,omitempty"`
	VesselId             int      `json:"VesselId,omitempty"`
	VesselName           string   `json:"VesselName,omitempty"`
	CategoryName         string   `json:"CategoryName,omitempty"`
	CategoryId           int      `json:"CategoryId,omitempty"`
	SupplierDistrictId1  int      `json:"SupplierDistrictId1,omitempty"`
	SupplierDistrictId2  int      `json:"SupplierDistrictId2,omitempty"`
	SupplierDistrictId3  int      `json:"SupplierDistrictId3,omitempty"`
	PurchaserDistrictId3 int      `json:"PurchaserDistrictId3,omitempty"`
	PurchaserDistrictId2 int      `json:"PurchaserDistrictId2,omitempty"`
	PurchaserDistrictId1 int      `json:"PurchaserDistrictId1,omitempty"`
	ProductId            int      `json:"ProductId,omitempty"`
	HsCode               int      `json:"HsCode,omitempty"`
}

type Product struct {
	ProductName string  `json:"pname"`
	ProId       int64   `json:"pid"`
	Count       float64 `json:"count"`
}

type Response struct {
	Error string      `json:"error,omitempty"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	List  interface{} `json:"list,omitempty"`
	Total int64       `json:"total"`
}

type TopTenProduct struct {
	ProductName string `json:"product_name"`
	Count       int64  `json:"count"`
}

type Relationship struct {
	ParentID    int64          `json:"parent_id"`
	ParentName  string         `json:"parent_name"`
	CompanyID   int64          `json:"company_id"`
	CompanyName string         `json:"company_name"`
	Partner     []Relationship `json:"partner"`
}

type Category struct {
	CategoryName string  `json:"cnameEn"`
	CategoryId   int     `json:"cid"`
	Value        float64 `json:"value"`
}
