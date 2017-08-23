package service

import "github.com/zhangweilun/tradeweb/model"

/**
*
* @author willian
* @created 2017-08-22 15:12
* @email 18702515157@163.com
**/

func GetDidNameByDid(districtId int64) string {
	district := model.District{
		Did: districtId,
	}
	db.Get(&district)
	return district.DnameEn
}

func GetBuyer(companyId int) *model.BusinessesNew {
	businessesNew := model.BusinessesNew{Id: int64(companyId)}
	db.Get(&businessesNew)
	company := model.DataCompany{BusinessesId: int(businessesNew.Id)}
	db.Get(&company)
	businessesNew.Url = company.CompanyWebsite
	businessesNew.LinkPhone = company.LinkPhone
	return &businessesNew

}

func GetSupplier(companyId int) *model.SuppliersNew {
	suppliersNew := model.SuppliersNew{Id: int64(companyId)}
	db.Get(&suppliersNew)
	company := model.DataCompany{SuppliersId: int(suppliersNew.Id)}
	db.Get(&company)
	suppliersNew.Url = company.CompanyWebsite
	suppliersNew.LinkPhone = company.LinkPhone
	return &suppliersNew
}
