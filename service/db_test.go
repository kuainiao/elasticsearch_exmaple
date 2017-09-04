package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/zhangweilun/tradeweb/model"
)

/**
*
* @author willian
* @created 2017-08-22 15:17
* @email 18702515157@163.com
**/

func TestGetDidNameByDid(t *testing.T) {
	did := GetDidNameByDid(3)
	fmt.Println(did)
}

func TestGetSupplier(t *testing.T) {
	supplier := GetSupplier(507834)
	fmt.Println(supplier)
}

func TestGetBuyer(t *testing.T) {
	buyer := GetBuyer(429222)
	fmt.Println(buyer.LinkPhone)
}

func TestGetBuyerContacts(t *testing.T) {
	contacts := GetBuyerContacts(896373)
	fmt.Println(contacts)
}

//SELECT d.dname_en name, count(b.id) as value,d.did,d.longitude,d.latitude
//FROM `suppliers_new` AS `b` LEFT  JOIN `district` AS `d` ON b.did_level1 = d.did
//WHERE b.id in (519197,554742,682318,682321,1081315,2151030,1111360,757763,2397174,682319) AND longitude != 0
//GROUP BY d.did
func TestGetCompanyDistrictInfo(t *testing.T) {
	info := GetCompanyDistrictInfo("519197,554742,682318,682321,1081315,2151030,1111360,757763,2397174,682319",
		1)
	fmt.Println(info)
}

func TestGetCompanyContacts(t *testing.T) {
	contacts, err, _ := GetCompanyContacts(2, 10, 0, 933070, "khs3UGawcs_vL_39TqZPJw")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(contacts)
}

func TestGetUserByCondition(t *testing.T) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte("gls123"))
	cipherStr := md5Ctx.Sum(nil)
	fmt.Println(hex.EncodeToString(cipherStr))
	users := model.Users{
		UserName: "admin",
		Passwd:   hex.EncodeToString(cipherStr),
	}
	count, _ := GetUserByCondition(&users)

	fmt.Println(count)
}

func TestGetMapInfo(t *testing.T) {
	maps := GetMapInfo(0, 2, 0, 0, 0, 0)
	fmt.Println(maps)
}

func TestGetAllDistrictName(t *testing.T) {
	name, err := GetAllDistrictName(0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(name)
}