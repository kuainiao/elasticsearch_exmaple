package service

import (
	"fmt"
	"strconv"

	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/util"
)

/**
*
* @author willian
* @created 2017-08-18 17:48
* @email 18702515157@163.com
**/
var db = xorm.MustDB("default")

//GetMapInfo 首页得到地图信息
func GetMapInfo(ietype, dateType, pid, dlevel, cid, did int) *[]model.MapInfo {
	var maps []model.MapInfo
	var disProduct string
	var sum string
	var andpid string
	var anddlevel string
	if dateType == 2 {
		disProduct = "statis_dis_product_1 sdc"
	} else if dateType == 0 {
		disProduct = "statis_dis_product sdc"
	} else if dateType == 1 {
		disProduct = "statis_dis_product_0 sdc"
	} else if dateType == 3 {
		disProduct = "statis_dis_product_2 sdc"
	}
	if ietype == 0 {
		sum = "SUM(sdc.busi_count) value,"
	} else {
		sum = "SUM(sdc.supp_count) value,"
	}
	if pid == 0 {
		andpid = "AND sdc.cid =" + strconv.Itoa(cid)
	} else if pid != 0 {
		andpid = "AND sdc.pid = " + strconv.Itoa(pid)
	}
	if dlevel == 0 {
		anddlevel = "AND dis.level = 1"
	} else if dlevel == 1 {
		anddlevel = "AND dis.level = 2 AND dis.pid =" + strconv.Itoa(did)
	} else if dlevel == 2 {
		anddlevel = "AND dis.level = 3 AND dis.pid =" + strconv.Itoa(did)
	}
	var sql = "SELECT dis.did did,dis.dname_en name," + sum + "dis.longitude longitude,dis.latitude latitude" +
		" FROM " + disProduct + " LEFT JOIN district dis ON sdc.did = dis.did" +
		" WHERE dis.longitude != 0 " +
		andpid + " " + anddlevel +
		" GROUP BY sdc.did ORDER BY value DESC;"
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

//0 得到所有国家 1得到所有省份 2得到所有市
func GetAllDistrictName(districtLevel int) *[]model.District {

	var result []model.District
	district := model.District{
		Level: districtLevel + 1,
	}
	db.Cols("dname_en", "did").Find(&result, district)
	return &result
}



func GetMapRelation(ietype, dateType, did int) *[]model.MapInfo {
	var maps []model.MapInfo
	var groupBy string
	var whereBy string
	var andBy string
	var columns string
	if ietype == 0 {
		groupBy = " suppliers_did_level1"
		whereBy = "sdc.businesses_did_level1 = " + strconv.Itoa(did)
		columns = "sdc.suppliers_did_level1 did,sdc.suppliers_country name,sdc.s_longitude longitude,sdc.s_latitude latitude,SUM(sdc.suppliers_count) value"
	} else {
		groupBy = "businesses_did_level1"
		whereBy = "sdc.suppliers_did_level1 = " + strconv.Itoa(did)
		columns = "sdc.businesses_did_level1 did,sdc.businesses_country name,sdc.b_longitude longitude, sdc.b_latitude latitude,SUM(sdc.businesses_count) value"
	}
	if dateType == 0 {
		andBy = " AND sdc.datetype = 3"
	} else if dateType == 1 {
		andBy = " AND sdc.datetype = 0"
	} else if dateType == 2 {
		andBy = " AND sdc.datetype = 1"
	} else if dateType == 3 {
		andBy = " AND sdc.datetype = 2"
	}
	var sql = "SELECT "+ columns +
		" FROM  statis_dist_company sdc " +
		" WHERE " + whereBy + " AND sdc.s_longitude != 0" + andBy +
		" GROUP BY " + groupBy + ";"
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

func GetMapClickInfo(did int64) *model.District {
	var district model.District
	district.Did = did
	db.Get(&district)
	return &district
}
