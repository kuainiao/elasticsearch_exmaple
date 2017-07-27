package main

import (
	"encoding/json"
	"fmt"
	"testing"
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

func TestJson(t *testing.T) {

	b := []byte(`{"end_date":"","country":"","destination_port":"","supplier":"","original_country":"","purchase":"","pro_desc":"s","start_date":""}`)
	//json str 转map
	var dat map[string]interface{}
	if err := json.Unmarshal(b, &dat); err == nil {
		fmt.Println("==============json str 转map=======================")
		fmt.Println(dat)
		fmt.Println(dat["host"])
	}
	//json str 转struct
	var config QueryParam
	if err := json.Unmarshal(b, &config); err == nil {
		fmt.Println("================json str 转struct==")
		fmt.Println(config)
		fmt.Println(config.ProDesc)
	}
	param := QueryParam{
		Supplier: "SS",
	}
	q, _ := json.Marshal(param)
	fmt.Println(string(q))
}
