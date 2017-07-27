package constants

import (

	"log"
	"os"
    elastic "gopkg.in/olivere/elastic.v5"
	"sync"
)

/**
* 
* @author willian
* @created 2017-07-27 17:44
* @email 18702515157@163.com  
**/

const (
	es_host = "192.168.1.114"
	//DateFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
)

var es *elastic.Client
var once sync.Once
//single
func Instance() *elastic.Client {
	once.Do(func() {
			client, _ := elastic.NewClient(
				elastic.SetURL("http://"+es_host+":9200"),
				elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
				elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
				elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]]", 0)))
			es = client
	})
	return es
}
