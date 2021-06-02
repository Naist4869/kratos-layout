package nosql

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// DBComponent 对需要数据库的业务模块的抽象
type DBComponent interface {
	Keys() map[string]*Spec
	Init() error
	Collections() map[string]*mongo.Collection
}
