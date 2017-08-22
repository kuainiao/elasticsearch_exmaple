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

