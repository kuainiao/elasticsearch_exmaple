package util

import (
	"log"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/zhangweilun/gor"
	"github.com/zhangweilun/goxmlpath"
)

/**
*
* @author willian
* @created 2017-07-26 18:01
* @email 18702515157@163.com
**/

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

//保留小数点后几位
func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

//return GoString's buffer slice(enable modify string)
func StringBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// convert b to string without copy
func BytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// returns &s[0], which is not allowed in go
func StringPointer(s string) unsafe.Pointer {
	p := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Pointer(p.Data)
}

// returns &b[0], which is not allowed in go
func BytesPointer(b []byte) unsafe.Pointer {
	p := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return unsafe.Pointer(p.Data)
}

func PutData(iter *goxmlpath.Path, page *goxmlpath.Node) []string {
	var result []string
	items := iter.Iter(page)
	for items.Next() {
		item := items.Node()
		result = append(result, item.String())
	}
	return result
}
func Zip(lists ...[]string) func() []string {
	zip := make([]string, len(lists))
	i := 0
	return func() []string {
		for j := range lists {
			if i >= len(lists[j]) {
				return nil
			}
			zip[j] = lists[j][i]
		}
		i++
		return zip
	}
}

var Words = func() []string {
	var result []string
	options := &gor.Request_options{
		UserAgent: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Mobile Safari/537.36",
	}

	res, err := gor.Get("http://www.thesaurus.com/browse/bike", options)
	page, err := goxmlpath.ParseHTML(res.RawResponse.Body)
	if err != nil {
		log.Fatal(err)
	}
	words := goxmlpath.MustCompile(`//*[@id="hider_synonyms"]/div[@class="result synstart"]/text()`)
	array_word := PutData(words, page)
	iter := Zip(array_word)
	for tuple := iter(); tuple != nil; tuple = iter() {
		result = append(result, tuple[0])
	}
	return result
}
