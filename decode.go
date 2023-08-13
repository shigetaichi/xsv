package xsv

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrEmptyCSVFile = errors.New("empty csv file given")
	ErrNoStructTags = errors.New("no csv struct tags found")
)

func mismatchStructFields(structInfo []fieldInfo, headers []string) []string {
	missing := make([]string, 0)
	if len(structInfo) == 0 {
		return missing
	}

	headerMap := make(map[string]struct{}, len(headers))
	for idx := range headers {
		headerMap[headers[idx]] = struct{}{}
	}

	for _, info := range structInfo {
		found := false
		for _, key := range info.keys {
			if _, ok := headerMap[key]; ok {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, info.keys...)
		}
	}
	return missing
}

func mismatchHeaderFields(structInfo []fieldInfo, headers []string) []string {
	missing := make([]string, 0)
	if len(headers) == 0 {
		return missing
	}

	keyMap := make(map[string]struct{})
	for _, info := range structInfo {
		for _, key := range info.keys {
			keyMap[key] = struct{}{}
		}
	}

	for _, header := range headers {
		if _, ok := keyMap[header]; !ok {
			missing = append(missing, header)
		}
	}
	return missing
}

func maybeMissingStructFields(structInfo []fieldInfo, headers []string) error {
	missing := mismatchStructFields(structInfo, headers)
	if len(missing) != 0 {
		return fmt.Errorf("found unmatched struct field with tags %v", missing)
	}
	return nil
}

// Check that no header name is repeated twice
func maybeDoubleHeaderNames(headers []string) error {
	headerMap := make(map[string]bool, len(headers))
	for _, v := range headers {
		if _, ok := headerMap[v]; ok {
			return fmt.Errorf("repeated header name: %v", v)
		}
		headerMap[v] = true
	}
	return nil
}

// Check if the outType is an array or a slice
func ensureOutType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the outInnerType is of type struct
func ensureOutInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func ensureOutCapacity(out *reflect.Value, csvLen int) error {
	switch out.Kind() {
	case reflect.Array:
		if out.Len() < csvLen-1 { // Array is not big enough to hold the CSV content (arrays are not addressable)
			return fmt.Errorf("array capacity problem: cannot store %d %s in %s", csvLen-1, out.Type().Elem().String(), out.Type().String())
		}
	case reflect.Slice:
		if !out.CanAddr() && out.Len() < csvLen-1 { // Slice is not big enough tho hold the CSV content and is not addressable
			return fmt.Errorf("slice capacity problem and is not addressable (did you forget &?)")
		} else if out.CanAddr() && out.Len() < csvLen-1 {
			out.Set(reflect.MakeSlice(out.Type(), csvLen-1, csvLen-1)) // Slice is not big enough, so grows it
		}
	}
	return nil
}

func getCSVFieldPosition(key string, structInfo *structInfo, curHeaderCount int) *fieldInfo {
	matchedFieldCount := 0
	for _, field := range structInfo.Fields {
		if field.matchesKey(key) {
			if matchedFieldCount >= curHeaderCount {
				return &field
			}
			matchedFieldCount++
		}
	}
	return nil
}

func createNewOutInner(outInnerWasPointer bool, outInnerType reflect.Type) reflect.Value {
	if outInnerWasPointer {
		return reflect.New(outInnerType)
	}
	return reflect.New(outInnerType).Elem()
}

func setInnerField(outInner *reflect.Value, outInnerWasPointer bool, index []int, value string, omitEmpty bool) error {
	oi := *outInner
	if outInnerWasPointer {
		// initialize nil pointer
		if oi.IsNil() {
			setField(oi, "", omitEmpty)
		}
		oi = outInner.Elem()
	}

	if oi.Kind() == reflect.Slice || oi.Kind() == reflect.Array {
		i := index[0]

		// grow slice when needed
		if i >= oi.Cap() {
			newcap := oi.Cap() + oi.Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newoi := reflect.MakeSlice(oi.Type(), oi.Len(), newcap)
			reflect.Copy(newoi, oi)
			oi.Set(newoi)
		}
		if i >= oi.Len() {
			oi.SetLen(i + 1)
		}

		item := oi.Index(i)
		if len(index) > 1 {
			return setInnerField(&item, false, index[1:], value, omitEmpty)
		}
		return setField(item, value, omitEmpty)
	}

	// because pointers can be nil need to recurse one index at a time and perform nil check
	if len(index) > 1 {
		nextField := oi.Field(index[0])
		return setInnerField(&nextField, nextField.Kind() == reflect.Ptr, index[1:], value, omitEmpty)
	}
	return setField(oi.FieldByIndex(index), value, omitEmpty)
}
