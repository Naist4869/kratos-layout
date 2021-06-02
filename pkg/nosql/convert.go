package nosql

import (
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

/*ObjectIDsToString 多个ObjectID转为字符串，id1,id2,id3,...,idx
参数:
*	ids	...primitive.ObjectID
返回值:
*	string	string
*/
func ObjectIDsToString(ids ...primitive.ObjectID) string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.Hex())
	}
	return strings.Join(result, ",")
}
