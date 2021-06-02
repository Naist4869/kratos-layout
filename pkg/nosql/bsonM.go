package nosql

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func PrettyBsonM(value bson.M) string {
	builder := &strings.Builder{}
	_, err := fmt.Fprint(builder, "{")
	i := 0
	for k, v := range value {
		if i != 0 {
			_, err = fmt.Fprintf(builder, ",")
		}
		if subValue, ok := v.(bson.M); ok {
			_, err = fmt.Fprintf(builder, "%s:%s", k, PrettyBsonM(subValue))

		} else {
			if _, ok = v.([]bson.M); ok {
				subValues := v.([]bson.M)
				subValueStrings := make([]string, 0, len(subValues))
				for _, subValue = range subValues {
					subValueStrings = append(subValueStrings, PrettyBsonM(subValue))
				}
				_, err = fmt.Fprintf(builder, "%s:%s", k, strings.Join(subValueStrings, ","))
			} else {
				_, err = fmt.Fprintf(builder, "%s:%v", k, v)
			}

		}

		i++
	}
	_, err = fmt.Fprint(builder, "}")
	if err != nil {

	}
	return builder.String()
}

/*CombineBsonM 合并多个Bson.M,如果有重复的字段，那么后出现会覆盖前者，返回的值都是经过复制的
参数:
*	documents	...bson.M
返回值:
*	bson.M	bson.M
*/
func CombineBsonM(documents ...bson.M) bson.M {
	count := 0
	for _, document := range documents {
		count += len(document)
	}

	result := make(bson.M, count)
	for _, document := range documents {
		for k, v := range document {
			result[k] = v
		}
	}
	return result

}
