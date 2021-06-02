package data

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-kratos/kratos-layout/pkg/nosql"

	"github.com/go-kratos/kratos-layout/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type greeterRepo struct {
	data        *Data
	log         *log.Helper
	collections map[string]*mongo.Collection // 数据表map

}

func (r *greeterRepo) Keys() map[string]*nosql.Spec {
	specs := make(map[string]*nosql.Spec, 3)
	{
		spec, err := nosql.NewSpec(biz.DBGreeterVersion, func() interface{} {
			return &biz.Greeter{}
		}, func(data interface{}) error {
			return nil
		})
		if err != nil {
			r.log.Errorf("构建updater失败: %+v", err)
		}
		specs[biz.DBGreeterKey] = spec
	}
	return specs
}

// Init 索引还在想
func (r *greeterRepo) Init() error {
	//return r.index()
	return nil
}

func (r *greeterRepo) index() error {
	return nosql.EnsureIndex(r.collections[biz.DBGreeterKey], []nosql.Index{
		{
			Name: "主键",
			Data: mongo.IndexModel{
				Keys:    bson.M{"_id": 1},
				Options: options.Index(),
			},
			Version: 1,
		},
	})
}

func (r *greeterRepo) Collections() map[string]*mongo.Collection {
	return r.collections
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
	return newGreeterRepo(data, logger)
}

func newGreeterRepo(data *Data, logger log.Logger) *greeterRepo {
	g := &greeterRepo{
		data:        data,
		log:         log.NewHelper(logger),
		collections: make(map[string]*mongo.Collection, 3),
	}
	if err := ComponentStart(g, data.mongodb); err != nil {
		g.log.Errorf("创建GreeterRepo失败: %+v", err)
	}
	return g
}

func (r *greeterRepo) CreateGreeter(ctx context.Context, g *biz.Greeter) error {
	return nil
}

func (r *greeterRepo) UpdateGreeter(ctx context.Context, g *biz.Greeter) error {
	return nil
}
