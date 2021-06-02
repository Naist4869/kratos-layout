package nosql

import "errors"

/* 定义了表格查询的一些方法，详情查看README.md#表格查询

 */
//TableRequest 表格查询参数
type TableRequest struct {
	Query map[string]interface{} `json:"query"` //查询条件
	Sort  []string               `json:"sort"`  //排序
	Start int                    `json:"start"` //开始位置，从0开始
	Limit int                    `json:"limit"` //本次最多数量
}

//Validate 验证
func (t TableRequest) Validate() error {
	if t.Limit < 0 {
		return errors.New("limit必须大于等于0")
	}
	if t.Start < 0 {
		return errors.New("start 必须大于等于0")
	}
	return nil
}

//TableResult 表格查询结果
type TableResult struct {
	Count  int         `json:"count"`  //总数量
	Result interface{} `json:"result"` //结果
}

/*NewTableResult 构建一个新的表格查询结果
参数:
*	count 	int			总数量
*	result	interface{}	结果，必须是slice
返回值:
*	TableResult	TableResult
*/
func NewTableResult(count int, result interface{}) TableResult {
	return TableResult{
		Count:  count,
		Result: result,
	}
}
