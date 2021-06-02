package nosql

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	versionKey = "meta.version"
)

// Spec 对数据库表的抽象描述
type Spec struct {
	collection *mongo.Collection
	version    int
	generator  func() interface{}
	updater    func(data interface{}) error
}

/*NewSpec 创建一个新的数据库版本规则
参数:
*	version   	int							// 期望版本，必须>1
*	generator 	func() interface{}			// 生成一个对应的对象，必须为指针
*	updater   	func(interface{}) error		// 升级方法，参数是所有版本低于version的数据，并且是调用generator()生成
返回值:
*	*Spec	*Spec
*	error	error
*/
func NewSpec(version int, generator func() interface{}, updater func(interface{}) error) (*Spec, error) {

	if version < 1 {
		return nil, errors.New("version参数必须大于等于1")
	}
	if generator == nil {
		return nil, errors.New("generator 参数不能为空")
	}
	if data := generator(); reflect.TypeOf(data).Kind() != reflect.Ptr || data == nil {
		return nil, errors.New("generator()方法必须返回一个非空指针")
	}
	if updater == nil {
		return nil, errors.New("updater参数不能为空")
	}
	return &Spec{
		version:   version,
		generator: generator,
		updater:   updater,
	}, nil
}
func (s *Spec) SetCollection(collection *mongo.Collection) {
	s.collection = collection
}

func (s *Spec) Update(key string) error {
	if s == nil {
		return nil
	}
	log.Printf("初始化升级数据,库名: %s", key)
	if hasHigher, err := s.hasHigherVersion(s.collection, s.version); err != nil {
		return errors.Wrap(err, "判断有无更高版本数据")
	} else {
		if hasHigher {
			return fmt.Errorf("数据库[%s]有高于版本的数据[%d]", key, s.version)
		}
		log.Printf("没有超过设计版本的数据,库名: %s", key)
		if err = s.updateLowerVersion(); err != nil {
			return errors.Wrap(err, "升级低版本数据")
		}
		return nil
	}
}

/*hasHigherVersion 判断有无更高版本的数据
参数:
*	collection	*mongo.Collection		数据库
*	version   	int						期望版本
返回值:
*	bool 	bool
*	error	error
*/
func (s *Spec) hasHigherVersion(collection *mongo.Collection, version int) (bool, error) {
	count, err := collection.CountDocuments(context.Background(), bson.M{versionKey: bson.M{"$gt": version}})
	return count > 0, err
}

/*updateLowerVersion 升级低版本数据
参数:
*	spec	Spec		数据定义
返回值:
*	error	error
*/
func (s *Spec) updateLowerVersion() error {
	if cursor, err := s.collection.Find(context.Background(), bson.M{"$or": []bson.M{
		{versionKey: bson.M{"$lt": s.version}},
		{versionKey: bson.M{"$exists": false}},
	}}); err != nil {
		return errors.Wrap(err, "查询错误")
	} else {
		bg := context.Background()

		defer cursor.Close(bg)
		for cursor.Next(bg) {
			record := s.generator()
			if err = cursor.Decode(record); err != nil {
				return errors.Wrap(err, "解码数据错误")
			}
			if err = s.updater(record); err != nil {
				return errors.Wrap(err, "升级版本错误")
			}
		}
	}
	return nil
}
