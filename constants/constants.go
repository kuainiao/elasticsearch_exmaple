package constants

import (
	"fmt"
	"log"
	"os"
	"sync"
	"github.com/henrylee2cn/faygo"
	"github.com/go-redis/redis"
	elastic "gopkg.in/olivere/elastic.v5"
)

/**
*
* @author willian
* @created 2017-07-27 17:44
* @email 18702515157@163.com
**/

const (
	Secret = "zhangweilun"

	esHost = "es.g2l-service.com"
	// esHost = "192.168.1.15:9200"
	//DateFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
	//ConfigDir config dir
	ConfigDir = "./config/"
	//ProductConfigFile product file name
	ProductConfigFile = "product.ini"
	IndexName         = "trade"
	TypeName          = "frank"
)

var (
	Stopwords = []string{
		"global",
		"a",
		"an",
		"and",
		"are",
		"as",
		"at",
		"be",
		"but",
		"by",
		"for",
		"if",
		"in",
		"into",
		"is",
		"it",
		"no",
		"not",
		"of",
		"on",
		"or",
		"such",
		"that",
		"the",
		"their",
		"then",
		"there",
		"these",
		"they",
		"this",
		"to",
		"was",
		"will",
		"with",
		"company",
		"gruop ",
		"service",
		"services",
		"inc",
		"inp",
		"llc",
		"corp",
		"ltd",
		"enterpris",
		"enterpre",
		"global",
		"llp",
	}
	es            *elastic.Client
	re            *redis.Client
	elasticSingle sync.Once
	redisSingle   sync.Once
	CategoryMap   = map[string]int{
		"Animal Products":                  1,
		"Vegetable Products":               2,
		"Animal and Vegetable Bi-Products": 3,
		"Foodstuffs":                       4,
		"Mineral Products":                 5,
		"Chemical Products":                6,
		"Plastics and Rubbers":             7,
		"Animal Hides":                     8,
		"Wood Products":                    9,
		"Paper Goods":                      10,
		"Textiles":                         11,
		"Footwear and Headwear":            12,
		"Stone And Glass":                  13,
		"Precious Metals":                  14,
		"Metals":                           15,
		"Machines":                         16,
		"Transportation":                   17,
		"Instruments":                      18,
		"Weapons":                          19,
		"Miscellaneous":                    20,
		"Arts and Antiques":                21,
		"Unspecified":                      22,
	}
	CategoryMapInt = map[int]string{
		1:  "Animal Products",
		2:  "Vegetable Products",
		3:  "Animal and Vegetable Bi-Products",
		4:  "Foodstuffs",
		5:  "Mineral Products",
		6:  "Chemical Products",
		7:  "Plastics and Rubbers",
		8:  "Animal Hides",
		9:  "Wood Products",
		10: "Paper Goods",
		11: "Textiles",
		12: "Footwear and Headwear",
		13: "Stone And Glass",
		14: "Precious Metals",
		15: "Metals",
		16: "Machines",
		17: "Transportation",
		18: "Instruments",
		19: "Weapons",
		20: "Miscellaneous",
		21: "Arts and Antiques",
		22: "Unspecified",
	}
	TopTenCategoryId = []int{
		20, 17, 7, 16, 11, 4, 15, 2, 13, 6,
	}
)

//ProductConfig product config struct
type ProductConfig struct {
	EsHost    string `ini:"es_host" comment:"es_host address"`
	UserName  string `ini:"user_name" comment:"es username"`
	PassWord  string `ini:"pass_word" comment:"es password"`
	Product   bool   `ini:"product" comment:"whether is the product"`
	RedisAddr string `ini:"redis_addr" comment:"redis_addr address"`
	RedisPass string `ini:"redis_pass" comment:"redis_pass "`
	RedisDb   int    `ini:"redis_db" comment:"redis_pass "`
	GlsUrl    string `ini:"gls_url" comment:"redis_pass "`
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

func Redis() *redis.Client {
	if re == nil {
		if Config.Product {
			redisSingle.Do(func() {
				re = redis.NewClient(&redis.Options{
					Addr:     Config.RedisAddr,
					Password: "",
					DB:       Config.RedisDb,
				})
			})
		} else {
			redisSingle.Do(func() {
				re = redis.NewClient(&redis.Options{
					Addr:     Config.RedisAddr,
					Password: Config.RedisPass,
					DB:       Config.RedisDb,
				})
			})
		}
	}
	return re
}

//Instance ...
func Instance() *elastic.Client {
	if es == nil {
		//生产
		if Config.Product {
			fmt.Println("product")
			elasticSingle.Do(func() {
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
			fmt.Println("develop")
			elasticSingle.Do(func() {
				client, err := elastic.NewClient(
					elastic.SetURL("http://"+Config.EsHost),
					elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
					// elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
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
