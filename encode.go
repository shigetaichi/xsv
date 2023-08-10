package xsv

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"io"
	"reflect"
	"sort"
)

var (
	ErrChannelIsClosed = errors.New("channel is closed")
)

type encoder struct {
	out io.Writer
}

func newEncoder(out io.Writer) *encoder {
	return &encoder{out}
}

func writeTo(writer CSVWriter, in interface{}, omitHeaders bool, removeFieldsIndexes []int, colIndex []int) error {
	colIndex = changeToSequence(colIndex)
	inValue, inType := getConcreteReflectValueAndType(in) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := ensureInType(inType); err != nil {
		return err
	}
	inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureInInnerType(inInnerType); err != nil {
		return err
	}

	inInnerStructInfo := getStructInfoNoCache(inInnerType)                                      // Get the inner struct info to get CSV annotations
	inInnerStructInfo.Fields = getFilteredFields(inInnerStructInfo.Fields, removeFieldsIndexes) // Filtered out ignoreFields from all fields

	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to Write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}
	csvHeadersLabels = reorderColumns(csvHeadersLabels, colIndex)
	if !omitHeaders {
		if err := writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	inLen := inValue.Len()
	for i := 0; i < inLen; i++ { // Iterate over container rows
		for j, fieldInfo := range inInnerStructInfo.Fields {
			csvHeadersLabels[j] = ""
			inInnerFieldValue, err := getInnerField(inValue.Index(i), inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			csvHeadersLabels[j] = inInnerFieldValue
		}
		csvHeadersLabels = reorderColumns(csvHeadersLabels, colIndex)
		if err := writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
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

func getFilteredFields(fields []fieldInfo, removeFieldsIndexes []int) []fieldInfo {
	var newFields []fieldInfo
	if len(removeFieldsIndexes) > 0 {
		for _, field := range fields {
			if !lo.Contains(removeFieldsIndexes, field.IndexChain[0]) {
				newFields = append(newFields, field)
			}
		}
	} else {
		newFields = fields
	}
	return newFields
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

/*
Make colIndex consist of sequential numbers starting from 0.
Ex. [1,2,5,8,0] -> [1,2,3,4,0]
*/
func changeToSequence(colIndex []int) []int {
	copiedColIndex := make([]int, len(colIndex))
	copy(copiedColIndex, colIndex)
	sort.Ints(copiedColIndex)

	for i, v := range colIndex {
		colIndex[i] = slices.Index(copiedColIndex, v)
	}
	return colIndex
}

func reorderColumns(row []string, colIndex []int) []string {
	newLine := make([]string, len(row))
	for from, to := range colIndex {
		newLine[to] = row[from]
	}
	return newLine
}
