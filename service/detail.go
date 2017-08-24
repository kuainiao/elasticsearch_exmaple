package service

import (
	"fmt"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/util"
	"strconv"
)

/**
*
* @author willian
* @created 2017-08-22 15:12
* @email 18702515157@163.com
**/
//GetDidNameByDid 得到地区名通过地区ID
func GetDidNameByDid(districtId int64) string {
	district := model.District{
		Did: districtId,
	}
	db.Get(&district)
	return district.DnameEn
}

//GetBuyer 通过id得到采购
func GetBuyer(companyId int) *model.BusinessesNew {
	businessesNew := model.BusinessesNew{Id: int64(companyId)}
	db.Get(&businessesNew)

	if businessesNew.Url == "" || businessesNew.LinkPhone == "" {
		company := model.DataCompany{BusinessesId: int(businessesNew.Id)}
		db.Get(&company)
		businessesNew.Url = company.CompanyWebsite
		businessesNew.LinkPhone = company.LinkPhone
		db.Update(businessesNew)
	}
	return &businessesNew

}

//GetBuyer 通过id得到供应商
func GetSupplier(companyId int) *model.SuppliersNew {
	suppliersNew := model.SuppliersNew{Id: int64(companyId)}
	db.Get(&suppliersNew)
	if suppliersNew.Url == "" || suppliersNew.LinkPhone == "" {
		company := model.DataCompany{SuppliersId: int(suppliersNew.Id)}
		db.Get(&company)
		suppliersNew.Url = company.CompanyWebsite
		suppliersNew.LinkPhone = company.LinkPhone
		db.Update(suppliersNew)
	}
	return &suppliersNew
}

//GetBuyerContacts 得到采购公司列表
func GetBuyerContacts(companyId int) *[]model.DataCompany {
	var dataCompany []model.DataCompany
	company := model.DataCompany{BusinessesId: companyId}
	db.Find(&dataCompany, company)
	return &dataCompany
}

//GetSupplierContacts 得到供应公司列表
func GetSupplierContacts(companyId int) *[]model.DataCompany {
	var dataCompany []model.DataCompany
	company := model.DataCompany{SuppliersId: companyId}
	db.Find(&dataCompany, company)
	return &dataCompany
}

//GetBuyerDistrictInfo 得到采购公司地区分布
func GetCompanyDistrictInfo(companyIds string, companyType int) *[]model.MapInfo {
	var maps []model.MapInfo
	var companyTableName string
	if companyType == 0 {
		companyTableName = "businesses_new"
	} else {
		companyTableName = "suppliers_new"
	}
	in := "(" + companyIds + ")"
	sql := "SELECT d.dname_en name, count(b.id) as value,d.did,d.longitude,d.latitude" +
		" FROM `" + companyTableName + "` AS `b` LEFT  JOIN `district` AS `d` ON b.did_level1 = d.did" +
		" WHERE b.id in " + in + " AND longitude != 0" +
		" GROUP BY d.did"
	slice, err := db.Query(sql)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(slice); i++ {
		info := model.MapInfo{}
		for k, v := range slice[i] {
			if k == "value" {
				atoi, _ := strconv.Atoi(util.BytesString(v))
				info.Value = atoi
			} else if k == "name" {
				info.Name = util.BytesString(v)
			} else if k == "did" {
				atoi, _ := strconv.Atoi(util.BytesString(v))
				info.Did = atoi
			} else if k == "longitude" {
				info.Longitude = util.BytesString(v)
			} else {
				info.Latitude = util.BytesString(v)
			}
		}
		maps = append(maps, info)
	}
	return &maps
}


//GetCompanyContacts 得到公司联系人
func GetCompanyContacts(companyType int, companyId int)  {
	
}