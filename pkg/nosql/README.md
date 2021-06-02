[toc]
# base

## 表格查询

定义统一的表格查询参数和表格查询结果，都在**table.go**中

### 查询参数

```go
type TableRequest struct {
	Query map[string]interface{} `json:"query"` //查询条件
	Sort  []string               `json:"sort"`  //排序
	Start int                    `json:"start"` //开始位置，从0开始
	Limit int                    `json:"limit"` //本次最多数量
}
```
对于前端而言，参数对应的结构为

```json
{
	"query":{
		"a":"5"
	},
	"sort":["-a","b","+c"],
	"start":0,
	"limit":50
}
```

*	query 表示查询条件，其中只包含了查询字段和值，对应的关系由前后端约定，比如上例中字段a的值为5，可以表示不等于5，也可以表示等于5
*	sort	字符串数组，每个字符串的结构为`-/+字段名`,-表示降序，+表示升序，如果没有-/+，那么表示对应的字段不参与排序。上例中的sort表示，a字段降序，b字段不参与，c字段升序
*	start	表示本次查询的开始位置，最小为0
*	limit	表示本次查询的最大数量


### 查询结果


```go
type TableResult struct {
	Count  int         `json:"count"`  //总数量
	Result interface{} `json:"result"` //结果
}
```

前端拿到的数据为

```json
{
	"count":1000,
	"result":[]
}

```

*	count 表示符合条件的总数量
*	result	表示具体结果