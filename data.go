package xsv

import (
	"fmt"
	"reflect"
)

type StructSlice[T any] []T

func (d StructSlice[T]) CheckIsStructSlice() error {
	if reflect.TypeOf(d).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("cannot use " + reflect.TypeOf(d).Elem().String() + ", only struct supported")
	}
	return nil
}
