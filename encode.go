package xsv

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"io"
	"reflect"
)

var (
	ErrChannelIsClosed = errors.New("channel is closed")
)

type encoder struct {
	out io.Writer
}

func ensureStructOrPtr(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Struct:
		fallthrough
	case reflect.Ptr:
		return nil
	}
	return fmt.Errorf("cannot use " + t.String() + ", only slice or array supported")
}

// Check if the inType is an array or a slice
func ensureInType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the inInnerType is of type struct
func ensureInInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func getInnerField(outInner reflect.Value, outInnerWasPointer bool, index []int) (string, error) {
	oi := outInner
	if outInnerWasPointer {
		if oi.IsNil() {
			return "", nil
		}
		oi = outInner.Elem()
	}

	if oi.Kind() == reflect.Slice || oi.Kind() == reflect.Array {
		i := index[0]

		if i >= oi.Len() {
			return "", nil
		}

		item := oi.Index(i)
		if len(index) > 1 {
			return getInnerField(item, false, index[1:])
		}
		return getFieldAsString(item)
	}

	// because pointers can be nil need to recurse one index at a time and perform nil check
	if len(index) > 1 {
		nextField := oi.Field(index[0])
		return getInnerField(nextField, nextField.Kind() == reflect.Ptr, index[1:])
	}
	return getFieldAsString(oi.FieldByIndex(index))
}

func getPickedFields(fields []fieldInfo, columnFieldsIndexes []int) []fieldInfo {
	var newFields []fieldInfo
	if len(columnFieldsIndexes) > 0 {
		for _, field := range fields {
			if lo.Contains(columnFieldsIndexes, field.IndexChain[0]) {
				newFields = append(newFields, field)
			}
		}
	} else {
		newFields = fields
	}
	return newFields
}
