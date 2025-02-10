package utils

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

func IsZeroType(value reflect.Value) bool {
	switch value.Kind() {
	// reject zero value for these types
	case reflect.String:
		return value.Len() == 0
	default:
		return false
	}
}

func UnmarshalBSON[T any](source any, dest T) error {
	b, err := bson.Marshal(source)
	if err != nil {
		return err
	}
	if err = bson.Unmarshal(b, &dest); err != nil {
		return err
	}
	return nil
}
