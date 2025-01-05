package utils

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

func IsZeroType(value reflect.Value) bool {
	zero := reflect.Zero(value.Type()).Interface()

	switch value.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		return value.Len() == 0
	default:
		return reflect.DeepEqual(zero, value.Interface())
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
