[TOC]
# 概述

mongo.go 文件包含了众多的基础mongo快捷操作

# 版本

## 当前版本

v0.1

## 版本历史

|   版本    |   修改人  |   修改时间    |   修改内容    |   备注    |
|---    |   --- |   --- |   --- |   --- |
|   v0.1    |   刘蕴唐  |  2019-07-09 14:43   |   初步完成    |   无  |


#  操作分类

目前支持以下操作

*   查询
*   索引


## 查询


### 基本查询

方法**Query**用于基本数据库查询

```go
Query(collection *mongo.Collection, ctx context.Context, query bson.M, sort []string, start, limit int64, include []string, exclude []string, data interface{}, all bool, collations ...*options.Collation) (result []interface{}, count int64, err error)
```

#### 参数说明

|   参数    |   含义    |
|   --- |   --- |
|   collection  |   数据库对象  |
|   ctx         |   上下文  |
|   query       |   查询条件    |
|   sort        |   排序规则    |
|   start       |   数据起点，从0开始   |
|   limit       |   返回数据数量，0表示不限制   |
|   include     |   返回值包含的字段名,数据库字段名 |
|   exluce      |   返回值排除的字段名,数据库字段名 |
|   data        |   数据库数据类型,必须是指针   |
|   all         |   是否返回全量计数(只包含查询条件)    |
|   collations  |   特殊的collation条件,最多只有一个    |


##### sort

sort 表示返回值的排序规则,格式为
>   +/-字段名

*   +表示增序
*   -表示降序

多个字段按照出现顺序，越早出现优先级越高

##### include/exclude

通常未必需要返回全部的字段，通过**包含/排除**双向描述

*   两者不能同时不为空(只能从一个方向指定)
*   对于_id字段，如果不需要，必须强制排除(mongodb规则)


##### data

受制于**go语法**,必须返回一个指针类型，才能反序列化


##### collation

collation 描述了特定的排序规则

#### 返回值

*   result []interface{}    查询结果，对于参数data类型*A,返回slice的每个元素都是A类型
*   count int64 如果all==true,返回全部数据数量,否则返回len(result)
*   err error   可能的错误

#### 使用技巧

通常在针对每种需要存储mongo的数据类型，定义一个最基础的Query

```go
func (s Server) Query(ctx context.Context, query bson.M, sort []string, start, limit int64, include []string, exclude []string, all bool) (cards []Card, count int64, err error) {
	var result []interface{}
	if len(sort) == 0 {
		sort = defaultSort
	}
	if result, count, err = base.Query(s.collection, ctx, query, sort, start, limit, include, exclude, Card{}, all, other.MoneyCollation); err != nil {
		return
	} else {
		cards = make([]Card, 0, len(result))
		for i := range result {
			cards = append(cards, result[i].(Card))
		}
		return
	}
}
```
主要包含以下内容:

*   将参数data,返回值result,改为具体的类型
*   添加默认的查询排序
*   添加默认的colllation

然后其他查询都是调用这个Query


### 构建查询条件

对于表格查询，通常我们需要重复、针对每个支持查询的字段做如下操作:

*   判断前端有无传值
*   如果没有，跳过
*   如果有，转换数据，将前端传入的数据转为数据库对应的类型，组合查询条件，进行查询

base定义了快捷、方便、统一的查询条件生成器，方便操作。


#### QuerySpec 和DbSpec

```go
// QuerySpec 查询条件
type QuerySpec map[string]DbSpec

// DbSpec 单个字段的查询条件
type DbSpec struct {
	Field   string   // 针对的字段
	Fields  []string // 针对的多个字段
	Op      string   // 操作符
	Convert Convert  // 转化规则
}
// Convert 转换规则,将字符串转为对应的类型
type Convert func(data string) (interface{}, error)
```

*   QuerySpec 定义了一个map, 外部数据->数据使用规则,key对应外部json数据的key
*   DbSpec 则定义了数据使用规则
    *   Field/Fields 该数据对应的数据库字段名(支持多个)
    *   Op              匹配操作符
    *   Convert         转换规则

#### Convert

为了方便日常处理，定义了几种常见的convert快捷方法

```go
/*TimeParse 时间解析，输出time.Time
参数:
*	data	string
返回值:
*	result	interface{}
*	err   	error
*/
func TimeParse(data string) (result interface{}, err error) {
	result, err = time.Parse("2006-01-02 15:04:05", data)
	return
}

/*Int64Parse int64解析
参数:
*	data	string
返回值:
*	result	interface{}
*	err   	error
*/
func Int64Parse(data string) (result interface{}, err error) {
	result, err = strconv.ParseInt(data, 10, 64)
	return
}

/*IntParse int解析
参数:
*	data	string
返回值:
*	result	interface{}
*	err   	error
*/
func IntParse(data string) (result interface{}, err error) {
	result, err = strconv.Atoi(data)
	return
}

/*RegexParse 正则表达式解析
参数:
*	data	string
返回值:
*	result	interface{}
*	err   	error
*/
func RegexParse(data string) (result interface{}, err error) {
	return ".*" + data + ".*", nil
}

/*ObjectIdParse ObjectID解析，输入为primitive.ObjectID
参数:
*	data	string
返回值:
*	result	interface{}
*	err   	error
*/
func ObjectIdParse(data string) (result interface{}, err error) {
	result, err = primitive.ObjectIDFromHex(data)
	return
}
```

#### BuildQuery

```go
/*BuildQuery 构建查询条件
参数:
*	data  	map[string]interface{}	传入的查询数据
*	specs 	QuerySpec			查询规则
*	strict	bool				是否严格，非严格时，将data中未在specs出现的字段转为普通$eq条件
返回值:
*	bson.M	bson.M
*	error 	error
*/
func BuildQuery(data map[string]interface{}, specs QuerySpec, strict bool) (bson.M, error)
```


##### 例子

```go
data := map[string]interface{}{
		"a":         1,
		"attend":    10,
		"attendMax": 20,
		"attendNo":  15,
		"c":         "a",
	}
	specs := QuerySpec{
		"attend": DbSpec{
			Op:    "$gt",
			Field: "attend",
		},
		"attendMax": DbSpec{
			Op:    "$lt",
			Field: "attend",
		},
		"attendNo": DbSpec{
			Op:     "$ne",
			Fields: []string{"attendNo1", "attendNo2"},
		},
		"c": {
			Op:    "$in",
			Field: "c",
			Convert: func(data string) (i interface{}, e error) {
				return []int{}, nil
			},
		},
	}
    query, err := BuildQuery(data, specs, false)
	t.Log(prettyBsonM(query))
```
输出结果如下
```json
{
    "attendNo1":{"$ne":15},
    "attendNo2":{"$ne":15},
    "attend":{"$gt":10,"$lt":20},
    "c":{"$in":[]},
    "a":1}
```

解析

*   外部数据包含了5个字段,a/atttend/attendMax/attendNo/c
*   specs定义了查询规则，针对attend/attendmax/attendNo/c字段进行处理
    *   attend字段
        *   对应数据库的*attend*字段
        *   操作符是$gt
        *   不需要进行转换
        *   结果为 **"attend":{"$gt":10}**
    *   attendMax字段
        *   对应数据库的*attend*字段
        *   操作符是$lt
        *   不需要进行转换
        *   结果为 **"attend":{"$lt":20}**
    *   attendNo字段
        *   对应数据库的*attendNo1 attendNo2*字段
        *   操作符是$ne
        *   不需要进行转换
        *   结果为 **attendNo1:{$ne:15},attendNo2：{$ne:15}**
    *   c 字段
        *   对应数据库 c字段
        *   操作符是$in
        *   转换是生成一个空的[]int
        *   结果为** "c":{"$in":[]}**
    *   a 字段
        *   规则中不存在
        *   BuildQuery调用时，使用了strict==false参数，所以自动以同字段、同类型加入条件
*   基于mongo查询语法，对条件进行组合，得到最终结果


#### BuildQueryWithLogic

对于常见的的判断条件，BuildQuery已经满足，但是mongo支持将多个条件进行逻辑组合，比如

*   同时满足条件组A，B
*   条件组A 由多个or组成
*   条件组B 由多个and 组成

从根本上分析，这种逻辑操作本身就是**巴科斯范式BNF**,表示的时候类似于**抽象语法树(AST)**,而我们原来的规范是无法满足这种结构的，所以定义了专用的规则


```go
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
func BuildQueryWithLogic(data map[string]interface{}, specs QuerySpec, strict bool, logic *LogicQuery) (bson.M, error)
```

##### LogicQueryType/LogicQuery
```go
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

```


##### 例子


```go
type queryData struct {
		A int `bson:"a"`
		B int `bson:"b"`
		C int `bson:"c"`
		D int `bson:"d"`
	}

	data := map[string]interface{}{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
	
	}
	specs := QuerySpec{
		"a": DbSpec{
			Op:      "$eq",
			Field:   "a",
			Convert: IntParse,
		},
		"b": DbSpec{
			Op:      "$eq",
			Field:   "b",
			Convert: IntParse,
		},
		"c": DbSpec{
			Op:      "$eq",
			Field:   "c",
			Convert: IntParse,
		},
		"d": DbSpec{
			Op:      "$eq",
			Field:   "d",
			Convert: IntParse,
		},
	}

	logic := &LogicQuery{
		Type: LogicOr,
		Fields: []LogicQuery{
			{
				Type: LogicAnd,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "a",
					},
					{
						Type: LogicNone,
						Key:  "b",
					},
				},
			},
			{
				Type: LogicAnd,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "c",
					},
					{
						Type: LogicNone,
						Key:  "d",
					},
				},
			},
		},
	}
	query, err := BuildQueryWithLogic(data, specs, true,logic)
	require.NoError(t, err)
	t.Log(prettyBsonM(query))
```

在普通的查询规则基础上，我们定义了对查询条件的组合

*	满足其中任何一组即可
	*	同时满足a,b两个字段的要求
	*	同时满足c,d两个字段的要求

生成的查询条件为
```json
{
	"$or":[
		{"$and":[
			{"a":{"$eq":1}},
			{"b":{"$eq":2}}
		]},
		{"$and":[
			{"c":{"$eq":3}},
			{"d":{"$eq":4}}
		]}
	]
}
```
测试如下

|	a	|	b	|	c	|d	|
|---	|	---	|	---	|	---	|
|	1	|	2	|	3	|	4	|
|	1	|	2	|	4	|	4	|
|	2	|	2	|	3	|	4	|
|	2	|	2	|	4	|	4	|

我们的规则为 a\==1且b\==2 或者 c\==3且d\==4
最终前3条数据查询得到
