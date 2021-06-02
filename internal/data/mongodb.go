package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos-layout/pkg/nosql"

	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoDB(conf *conf.Data, l log.Logger) (db *mongo.Database, cf func(), err error) {
	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      conf.Mongodb.Username,
		Password:      conf.Mongodb.Password,
		AuthSource:    conf.Mongodb.AuthSource,
	}
	//client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)).SetAuth(auth).SetRegistry(bson.NewRegistryBuilder().RegisterCodec(reflect.TypeOf((*orderModel.MessageKind)(nil)).Elem(), orderModel.MessageKindCodec{}).RegisterCodec(reflect.TypeOf((*messageModel.Kind)(nil)).Elem(), messageModel.Codec{}).Build()))

	client, err := mongo.Connect(context.Background(), options.Client().SetHosts(conf.Mongodb.Hosts).SetReplicaSet("replicaset").SetReadPreference(readpref.Primary()).SetAuth(auth).SetRegistry(bson.NewRegistryBuilder().Build()))
	if err != nil {
		return
	}
	cf = func() {
		l.Log(log.LevelInfo, "closing the mongodb resources")
		if err := client.Disconnect(context.Background()); err != nil {
			l.Log(log.LevelError, "关闭Mongo客户端连接池失败: %#v", err)
		}
	}
	// 这里可以root登录访问别的db  但是目前只使用一个数据库  可能访问oplog的时候需要访问admin数据库
	db = client.Database(conf.Mongodb.AuthSource)
	return
}

func ComponentStart(component nosql.DBComponent, client *mongo.Database) (err error) {
	keys := component.Keys()
	for key, spec := range keys {
		collection := client.Collection(key)
		component.Collections()[key] = collection
		if spec != nil {
			spec.SetCollection(collection)
		}
	}
	for key, spec := range keys {
		if spec != nil {
			if err = component.Init(); err != nil {
				return fmt.Errorf("初始化%s component失败: %w", key, err)
			}
			if err := spec.Update(key); err != nil {
				return fmt.Errorf("检查并升级%s collection数据失败: %w", key, err)
			}
		}
	}
	return
}
