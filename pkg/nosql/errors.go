package nosql

import "strings"

func IsInsertDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "E11000 duplicate key error collection")
}

func IsUpdateLeastError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "update document must have at least one element")
}
