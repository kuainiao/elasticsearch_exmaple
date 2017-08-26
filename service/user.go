package service

import "github.com/zhangweilun/tradeweb/model"

/**
*
* @author willian
* @created 2017-08-25 15:39
* @email 18702515157@163.com
**/

//通过条件查询用户
func GetUserByCondition(users *model.Users) (count int64, err error) {
	count, err = db.Count(users)
	return count, err
}
