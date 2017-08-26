package util

import (
	"testing"
	"fmt"
)

/**
* 
* @author willian
* @created 2017-08-24 15:19
* @email 18702515157@163.com  
**/

func TestTrimFrontBack(t *testing.T) {
	s := " sofa bed "
	back := TrimFrontBack(s)
	fmt.Println(back)
}