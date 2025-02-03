package utils

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

func IsZeroType(value reflect.Value) bool {
	zero := reflect.Zero(value.Type()).Interface()

	switch value.Kind() {
	// reject zero value for these types
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		return value.Len() == 0
	// accept zero value for these types
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Bool:
		return false
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
