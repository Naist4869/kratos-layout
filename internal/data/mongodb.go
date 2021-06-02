package data

import (
	"context"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoDB(conf *conf.Data, l log.Logger) (client *mongo.Client, cf func(), err error) {
	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      conf.Mongodb.Username,
		Password:      conf.Mongodb.Password,
		AuthSource:    conf.Mongodb.AuthSource,
	}
	//client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)).SetAuth(auth).SetRegistry(bson.NewRegistryBuilder().RegisterCodec(reflect.TypeOf((*orderModel.MessageKind)(nil)).Elem(), orderModel.MessageKindCodec{}).RegisterCodec(reflect.TypeOf((*messageModel.Kind)(nil)).Elem(), messageModel.Codec{}).Build()))

	client, err = mongo.Connect(context.Background(), options.Client().SetHosts(conf.Mongodb.Hosts).SetReplicaSet("replicaset").SetReadPreference(readpref.Primary()).SetAuth(auth).SetRegistry(bson.NewRegistryBuilder().Build()))
	if err != nil {
		return
	}
	cf = func() {
		l.Log(log.LevelInfo, "closing the mongodb resources")
		if err := client.Disconnect(context.Background()); err != nil {
			l.Log(log.LevelError, "关闭Mongo客户端连接池失败: %#v", err)
		}
	}

	return
}
