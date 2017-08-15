package constants

import (
	"log"
	"os"
	"sync"

	"github.com/henrylee2cn/faygo"

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
	// esHost = "192.168.1.15:9200"
	//DateFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
	//ConfigDir config dir
	ConfigDir = "./config/"
	//ProductConfigFile product file name
	ProductConfigFile = "product.ini"
)

var es *elastic.Client
var once sync.Once

//ProductConfig product config struct
type ProductConfig struct {
	EsHost   string `ini:"es_host" comment:"es_host address"`
	UserName string `ini:"user_name" comment:"es username"`
	PassWord string `ini:"pass_word" comment:"es password"`
	Product  bool   `ini:"product" comment:"whether is the product"`
}

//Config Product Config
var Config = func() ProductConfig {
	var productConfig = ProductConfig{}
	filename := ConfigDir + ProductConfigFile
	err := faygo.SyncINI(
		&productConfig,
		func(onecUpdateFunc func() error) error {
			return onecUpdateFunc()
		},
		filename,
	)
	if err != nil {
		panic(err)
	}
	return productConfig
}()

//Instance ...
func Instance() *elastic.Client {
	if es == nil {
		//生产
		if Config.Product {
			once.Do(func() {
				client, err := elastic.NewClient(
					elastic.SetURL("http://"+Config.EsHost),
					elastic.SetSniff(false),
				)
				if err != nil {
					panic(err)
				}
				es = client
			})
		} else {
			once.Do(func() {
				client, err := elastic.NewClient(
					elastic.SetURL("http://"+Config.EsHost),
					elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
					elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
					elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]]", 0)),
					elastic.SetBasicAuth(Config.UserName, Config.PassWord),
					elastic.SetSniff(false),
				)
				if err != nil {
					panic(err)
				}
				es = client
			})
		}

	}
	return es
}
