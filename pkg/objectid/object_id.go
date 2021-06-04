package objectid

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type ObjectID string

func (o *ObjectID) UnmarshalBSONValue(bt bsontype.Type, v []byte) error {
	value := bsoncore.Value{
		Type: bt,
		Data: v,
	}
	if value, ok := value.ObjectIDOK(); !ok {
		return fmt.Errorf("错误的类型[%s],不是ObjectID", v)
	} else {
		*o = ObjectID(value.Hex())
		return nil
	}
}
func (o ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	objID, err := primitive.ObjectIDFromHex(string(o))
	if err != nil {
		return bsontype.Null, nil, err
	}
	return bsontype.ObjectID, objID[:], nil
}
