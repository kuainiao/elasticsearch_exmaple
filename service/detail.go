package service

import (
	"fmt"
	"strconv"

	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/util"
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
	db.Cols("dname_en").Get(&district)
	return district.DnameEn
}

//GetBuyer 通过id得到采购
func GetBuyer(companyId int) *model.BusinessesNew {
	businessesNew := model.BusinessesNew{Id: int64(companyId)}
	db.Omit("create_time","update_time","name").Get(&businessesNew)

	if businessesNew.Url == "" || businessesNew.LinkPhone == "" {
		company := model.DataCompany{BusinessesId: int(businessesNew.Id)}
		db.Get(&company)
		businessesNew.Url = company.CompanyWebsite
		businessesNew.LinkPhone = company.LinkPhone
		//buyerConfition := model.BusinessesNew{Id: int64(companyId)}
		//db.Update(businessesNew, buyerConfition)
	}
	return &businessesNew

}

//GetBuyer 通过id得到供应商
func GetSupplier(companyId int) *model.SuppliersNew {
	suppliersNew := model.SuppliersNew{Id: int64(companyId)}
	db.Get(&suppliersNew)
	if suppliersNew.Url == "" || suppliersNew.LinkPhone == "" {
		company := model.DataCompany{SuppliersId: int(suppliersNew.Id)}
		db.Omit("create_time","update_time","name").Get(&company)
		suppliersNew.Url = company.CompanyWebsite
		suppliersNew.LinkPhone = company.LinkPhone
		//supplierConfition := model.SuppliersNew{Id: int64(companyId)}
		//db.Update(suppliersNew, supplierConfition)
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
func GetCompanyContacts(pageNo, pageSize, companyType, companyId int, guid string) (*[]model.Contact, error, int64) {
	var contacts []model.Contact
	var total int64
	if companyType == 0 {
		buyer := model.BusinessesNew{Id: int64(companyId)}
		db.Cols("did_level1").Get(&buyer)
		contact := model.Contact{BusinessesId: int64(companyId)}
		count, err := db.Count(&contact)
		if count == 0 {
			return nil, err, 0
		}
		total = count
		if buyer.DidLevel1 == 1 {
			//中国
			db.Omit("email", "tel_phone", "other_link").Get(&contact)
			contacts = append(contacts, contact)
		} else {
			//不为中国
			start := (pageNo - 1) * pageSize
			db.Limit(pageSize, start).Find(&contacts, contact)
		}
	} else {
		supplier := model.SuppliersNew{Id: int64(companyId)}
		db.Cols("did_level1").Get(&supplier)
		contact := model.Contact{SuppliersId: int64(companyId)}
		count, err := db.Count(&contact)
		if count == 0 {
			return nil, err, 0
		}
		total = count
		if supplier.DidLevel1 != 1 {
			//不为中国
			db.Omit("email", "tel_phone", "other_link").Get(&contact)
			contacts = append(contacts, contact)
		} else {
			//中国
			start := (pageNo - 1) * pageSize
			db.Limit(pageSize, start).Cols("country", "city", "other_link", "sex",
				"name", "mobile", "id", "position", "tel_phone", "depar_name", "email").Find(&contacts, contact)
		}
	}
	return &contacts, nil, total
}
