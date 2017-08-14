package constants

import (
	"log"
	"os"
	"sync"

	elastic "gopkg.in/olivere/elastic.v5"
)

/**
*
* @author willian
* @created 2017-07-27 17:44
* @email 18702515157@163.com
**/

const (
	esHost = "es.g2l-service.com"
	//esHost = "192.168.1.114:9200"
	//DateFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
)

var es *elastic.Client
var once sync.Once

//Instance ...
func Instance() *elastic.Client {
	if es == nil {
		once.Do(func() {
			client, err := elastic.NewClient(
				elastic.SetURL("http://"+esHost),
				elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
				elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
				// elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]]", 0)),
				elastic.SetBasicAuth("admin", "4Dm1n.3s"),
				elastic.SetSniff(false),
			)
			if err != nil {
				panic(err)
			}
			es = client
		})
	}
	return es
}
