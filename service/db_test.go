package service

import (
	"testing"
	"fmt"
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
