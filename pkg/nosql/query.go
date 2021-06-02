package nosql

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MakeSelect(include, exclude []string) (selection bson.M, err error) {
	if err = validateSelect(include, exclude); err != nil {
		return
	}
	selection = makeSelect(include, exclude)
	return
}

func validateSelect(include, exclude []string) (err error) {
	if len(include) != 0 && len(exclude) != 0 {
		err = errors.New("两个参数必须至少有一个为空")
		return
	}
	return
}

func makeSelect(include, exclude []string) (fields bson.M) {
	fields = make(bson.M, 2)
	for _, field := range include {
		fields[field] = 1
	}
	for _, field := range exclude {
		fields[field] = 0
	}
	return
}
func convertSort(sorts []string) (result bson.D) {
	result = make([]bson.E, 0, len(sorts))
	value := 1
	for _, sort := range sorts {
		if sort[:1] == "-" {
			sort = sort[1:]
			value = -1
		}
		result = append(result, bson.E{
			Key:   sort,
			Value: value,
		})
	}
	return
}

func BaseQuery(collection *mongo.Collection, ctx context.Context, query bson.M, sort []string, start, limit int64, include []string, exclude []string, data interface{}, collations ...*options.Collation) (result []interface{}, count int64, err error) {
	var selection bson.M
	if selection, err = MakeSelect(include, exclude); len(selection) == 0 {
		selection = nil
	}
	option := &options.FindOptions{
		Sort:       convertSort(sort),
		Limit:      &limit,
		Skip:       &start,
		Projection: selection,
	}
	switch len(collations) {
	case 0:
	case 1:
		option.Collation = collations[0]
	default:
		err = errors.New("collations参数最多只有1个")
		return
	}
	var cursor *mongo.Cursor
	if cursor, err = collection.Find(ctx, query, option); err != nil {
		err = fmt.Errorf("find驱动:%w", err)
		return
	}
	t := reflect.TypeOf(data)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		element := reflect.New(t)
		if err = cursor.Decode(element.Interface()); err != nil {
			err = fmt.Errorf("bson解码:%w", err)
			return
		}
		result = append(result, element.Elem().Interface())
		count++
	}
	return
}

// ---------------------------------------------多条件组装查询---------------------------------------------------------------

type querySpec struct {
	key     string
	op      string
	convert Convert
}

// Convert 转换规则,将字符串转为对应的类型
type Convert func(data string) (interface{}, error)

/*BuildQuery 构建查询条件
参数:
*	data  	map[string]interface{}	传入的查询数据
*	specs 	QuerySpec			查询规则
*	strict	bool				是否严格，非严格时，将data中未在specs出现的字段转为普通$eq条件
返回值:
*	bson.M	bson.M
*	error 	error
*/
func BuildQuery(data map[string]interface{}, specs QuerySpec, strict bool) (bson.M, error) {
	result := make(map[string]interface{}) // 结果
	mappedSpec, dynamic := remapQuerySpecByField(specs)
	used := make(map[string]bool) // 记录使用过的数据
	var err error
	for field, spec := range mappedSpec { // 遍历规则
		for _, s := range spec { // 单条规则
			if value, exist := data[s.key]; exist { // 如果规则能够适用

				if s.convert != nil {
					if value, err = s.convert(value.(string)); err != nil {
						return nil, fmt.Errorf("转换数据发生问题:\n\t数据:%#v\n\t类型:%T\n\t字段:%s", data[s.key], data[s.key], s.key)
					}
				} else {
				}

				if _, exist = result[field]; !exist {
					result[field] = make(map[string]interface{})
				}
				if len(spec) > 1 {
					result[field].(map[string]interface{})[s.op] = value
					if s.op == "$regex" {
						result[field].(map[string]interface{})["$options"] = "i"
					}
				} else if len(spec) == 1 {
					switch s.op {
					case "$regex":
						result[field] = bson.M{s.op: value, "$options": "i"}
					case "$in", "$nin":
						if value != nil {
							reflectValue := reflect.ValueOf(value)
							if reflectValue.Kind() != reflect.Slice {
								return nil, fmt.Errorf("$in值必须是slice,对应字段[%s]", s.key)
							}
							if !reflectValue.IsNil() {
								result[field] = bson.M{s.op: value}
							}
						}
					default:
						result[field] = bson.M{s.op: value}
					}

				} else {
					delete(result, field)
				}
				used[s.key] = true // 记录已经使用过的数据，用过的数据不能直接删除，可能会多次使用
			}
		}
	}
	for key, convert := range dynamic {
		if value, exist := data[key]; exist { //外部传入值
			if value, err = convert(value.(string)); err != nil {
				return nil, fmt.Errorf("转换数据发生问题:\n\t数据:%#v\n\t类型:%T\n\t字段:%s", data[key], reflect.TypeOf(data[key]).String(), key)
			} else {
				//todo: 选择一种更好的方法， 类型断言失败
				if query, ok := value.(primitive.M); !ok {
					return nil, fmt.Errorf("动态字段[%s]Convert第一个返回值必须是bson.M,实际上是[%s]", key, reflect.TypeOf(value).String())
				} else {
					for key, condition := range query {
						switch key {
						case "$or", "$and":
							//todo: 如果有多个$or或者$and 需要汇总
							result[key] = condition
						default:
							if result[key] == nil {
								result[key] = make(map[string]interface{})
							}
							if fullCondition, ok := condition.(primitive.M); ok {
								for k, v := range fullCondition {
									result[key].(map[string]interface{})[k] = v
								}
							} else {
								spew.Dump(result[key], result, key)
								switch t := result[key].(type) {
								case primitive.M:
									t["$eq"] = condition
								case map[string]interface{}:
									result[key].(map[string]interface{})["$eq"] = condition
								}

							}
						}

					}

				}

			}
			delete(data, key)
		}
	}
	if !strict { // 如果不是严厉规则，将其余的数据加上，判断是否相等
		for k, v := range data {
			if !used[k] {
				result[k] = v
			}

		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	query := clean(bson.M(result))
	return query, nil
}

func clean(query bson.M) bson.M {
	deleted := make([]string, 0, len(query))
	for k, v := range query {
		if value, ok := v.(map[string]interface{}); ok {
			if len(value) == 0 {
				deleted = append(deleted, k)
			}
		}
	}
	for _, key := range deleted {
		delete(query, key)
	}
	return query
}

// QuerySpec 查询条件，外部数据->数据使用规则
type QuerySpec map[string]DbSpec

// DbSpec 数据使用规则
type DbSpec struct {
	Field   string   // 针对的字段
	Fields  []string // 针对的多个字段
	Dynamic bool     // 是否动态,结果是条件bson.M，而不是值
	Op      string   // 操作符
	Convert Convert  // 转化规则
}

func remapQuerySpecByField(spec QuerySpec) (ordinary map[string][]querySpec, dynamic map[string]Convert) {
	ordinary = make(map[string][]querySpec)
	for k, v := range spec {
		if v.Dynamic { //动态字段
			if len(dynamic) == 0 {
				dynamic = make(map[string]Convert, 10)
			}
			dynamic[k] = v.Convert
			continue
		}
		if v.Field != "" {
			if _, exist := ordinary[v.Field]; !exist {
				ordinary[v.Field] = make([]querySpec, 0, 2)
			}
		} else {
			for _, field := range v.Fields {
				if _, exist := ordinary[field]; !exist {
					ordinary[field] = make([]querySpec, 0, 2)
				}
			}
		}
		if v.Field != "" {
			ordinary[v.Field] = append(ordinary[v.Field], querySpec{key: k, op: v.Op, convert: v.Convert})
		} else {
			for _, field := range v.Fields {
				ordinary[field] = append(ordinary[field], querySpec{key: k, op: v.Op, convert: v.Convert})
			}
		}
	}
	return
}

/*BuildQueryWithLogic 构建查询，并且支持逻辑操作
参数:
*	data  	map[string]interface{}		输入数据
*	specs 	QuerySpec					查询规则
*	strict	bool						是否严格(查看BuildQuery)
*	logic 	*LogicQuery					逻辑条件
返回值:
*	bson.M	bson.M						生成的查询条件
*	error 	error						可能的错误
*/
func BuildQueryWithLogic(data map[string]interface{}, specs QuerySpec, strict bool, logic *LogicQuery) (bson.M, error) {
	query, err := BuildQuery(data, specs, strict)
	if err != nil {
		return nil, err
	}
	return generate(logic, query), nil
}

// LogicQueryType 逻辑条件类型
type LogicQueryType int

const (
	// LogicNone 无逻辑条件，最后的数据
	LogicNone LogicQueryType = iota
	// LogicAnd	逻辑与
	LogicAnd
	// LogicOr	逻辑或
	LogicOr
)

// LogicQuery 逻辑查询条件
type LogicQuery struct {
	Type   LogicQueryType // 逻辑类型,如果为LogicOr时，Key不能为空,如果为其他类型时,Fields不能为空
	Fields []LogicQuery   // 内部成员
	Key    string         // 字段名
}

func generate(logic *LogicQuery, query bson.M) bson.M {
	switch logic.Type {
	case LogicAnd, LogicOr:

		return generateLogic(logic, query)
	case LogicNone:
		return query
	default:
		return query
	}
}

func generateLogic(logic *LogicQuery, query bson.M) bson.M {
	if logic.Type == LogicNone {
		return bson.M{logic.Key: query[logic.Key]}
	}
	fields := make([]bson.M, 0, len(logic.Fields))
	for _, field := range logic.Fields {
		fieldQuery := generateLogic(&field, query)
		fields = append(fields, fieldQuery)

	}
	operator := "$and"
	if logic.Type == LogicOr {
		operator = "$or"
	}
	return bson.M{operator: fields}
}
