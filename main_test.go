package main

import (
	"encoding/json"
	"fmt"


	"testing"
	"time"

	"net/url"
)

/**
*
* @author willian
* @created 2017-07-27 17:20
* @email 18702515157@163.com
**/
type QueryParam struct {
	Country         string `json:"country"`
	ProDesc         string `json:"pro_desc"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	Purchase        string `json:"purchase"`
	Supplier        string `json:"supplier"`
	OriginalCountry string `json:"original_country"`
	DestinationPort string `json:"destination_port"`
}

//type jsonTime time.Time
//
////实现它的json序列化方法
//func (this jsonTime) MarshalJSON() ([]byte, error) {
//	fmt.Println(this)
//	var stamp = fmt.Sprintf("\"%s\"", time.Time(this).Format("2006-01-02"))
//	return []byte(stamp), nil
//}

type Frank struct {
	OrderId          int64   `json:"OrderId"`
	MudiPort         string  `json:"MudiPort"`
	Purchaser        string  `json:"Purchaser"`
	PurchaserAddress string  `json:"PurchaserAddress"`
	SupplierId       int64   `json:"SupplierId"`
	PurchaserId      int64   `json:"PurchaserId"`
	OrderNo          string  `json:"OrderNo"`
	OrderWeight      float64 `json:"OrderWeight"`
	Supplier         string  `json:"Supplier"`
	ProDesc          string  `json:"ProDesc"`
	OriginalCountry  string  `json:"OriginalCountry"`
	SupplierAddress  string  `json:"SupplierAddress"`
	FrankTime        JsonDate    `json:"FrankTime"`
	OrderVolume      float64 `json:"OrderVolume"`
	QiyunPort        string  `json:"QiyunPort"`
	TradeNumber      int64   `json:"TradeNumber"`
}

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

func TestScanf(t *testing.T) {
	str := "2016-09-23"
	var (
		year int
		mon  int
		mday int
	)
	fmt.Sscanf(str, "%d-%02d-%02d", &year, &mon, &mday)
	fmt.Println(year)
}

func TestPath(t *testing.T) {
	unescape, _ := url.PathUnescape("vacuum%20cleaner")
	fmt.Println(unescape)
}

func TestJson(t *testing.T) {

	b := []byte(`{"OrderId": 22829801, "MudiPort": "3002, TACOMA, WA", "Purchaser": "563215 BC LTD", "PurchaserAddress": "5930 NO 6 RD UNIT 325 RICHMOND BC V6V1Z1 CA", "SupplierId": 506056, "PurchaserId": 80, "OrderNo": "FTNVTPS000040583", "OrderWeight": 0.065, "Supplier": "SPANK INDUSTRIES CO LTD", "ProDesc": "BIKE PARTSON BOARD DATE<br/>", "OriginalCountry": "TW, TAIWAN", "SupplierAddress": "5F NO 62 JHONGMING S RD TAICHUNG TW", "FrankTime": "2015-07-03", "OrderVolume": 1.0, "QiyunPort": "58309, KAO HSIUNG", "TradeNumber": 0}`)
	//json str 转map
	//var dat map[string]interface{}
	//if err := json.Unmarshal(b, &dat); err == nil {
	//	fmt.Println("==============json str 转map=======================")
	//	fmt.Println(dat)
	//	fmt.Println(dat["host"])
	//}
	//json str 转struct
	var config Frank
	if err := json.Unmarshal(b, &config); err == nil {
		fmt.Println("================json str 转struct==")
		fmt.Println(config)
		fmt.Println(config.ProDesc)
	} else {
		fmt.Println(err)
	}
	param := Frank{
		Supplier:  "SS",
		FrankTime: config.FrankTime,
	}
	fmt.Println(param)
	q, err := json.Marshal(&param)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(q))
}

//func TestDiv(t *testing.T) {
//	encodeurl:= url.QueryEscape("5foGXY1Jqys5tQQCSkDF/dk+KxE=")
//	fmt.Println(encodeurl)
//	unescape, queryUnescape := url.QueryUnescape(encodeurl)
//}