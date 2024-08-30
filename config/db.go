package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Db *mongo.Database
)

// 使用时解除注释
// func init() {
// 	var err error
// 	opt := options.Client()
// 	opt.Hosts = Conf.Db.Mongo.Hosts         //主机地址数组
// 	opt.SetLocalThreshold(time.Second * 3). //只使用与mongo操作耗时小于3秒的
// 						SetMaxConnIdleTime(5 * time.Millisecond). //指定连接可以保持空闲的最大毫秒数
// 						SetMaxPoolSize(200)                       //使用最大的连接数

// 	wc := writeconcern.New(writeconcern.WMajority())
// 	readconcern.Majority()
// 	opt.ReadConcern = readconcern.Majority()
// 	opt.WriteConcern = wc
// 	var client *mongo.Client
// 	if client, err = mongo.Connect(getContext(), opt); err != nil {
// 		checkErr(err)
// 	}
// 	Db = client.Database(Conf.Db.Mongo.Database)
// }

func checkErr(err error) {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("没有查到数据")
			os.Exit(0)
		} else {
			fmt.Println(err)
			os.Exit(0)
		}

	}
}

func getContext() (ctx context.Context) {
	ctx, err := context.WithTimeout(context.Background(), 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	return
}
