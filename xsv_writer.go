package xsv

import (
	"encoding/csv"
	"reflect"
)

type XsvWriter[T any] struct {
	XsvWrite[T]
	writer *csv.Writer
}

func NewXsvWriter[T any](xsvWrite XsvWrite[T]) *XsvWriter[T] {
	return &XsvWriter[T]{XsvWrite: xsvWrite}
}

func (xw *XsvWriter[T]) Comma(comma rune) *XsvWriter[T] {
	xw.writer.Comma = comma
	return xw
}

func (xw *XsvWriter[T]) UseCRLF(useCRLF bool) *XsvWriter[T] {
	xw.writer.UseCRLF = useCRLF
	return xw
}

func (xw *XsvWriter[T]) Write(data []T) error {
	inValue, inType := getConcreteReflectValueAndType(data) // Get the concrete type (not pointer) (Slice<?> or Array<?>)

	inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureInInnerType(inInnerType); err != nil {
		return err
	}

	fieldInfos := getFieldInfos(inInnerType, []int{}, []string{}, xw.TagName, xw.TagSeparator, xw.nameNormalizer) // Get the inner struct info to get CSV annotations
	fieldInfos = xw.getSelectedFieldInfos(fieldInfos)
	if err := xw.checkSortOrderSlice(len(fieldInfos)); err != nil {
		return err
	}
	fieldInfos = reorderColumns[fieldInfo](fieldInfos, xw.SortOrder)
	inInnerStructInfo := &structInfo{fieldInfos}

	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
		if newHeader, ok := xw.HeaderModifier[fieldInfo.getFirstKey()]; ok { // modify header name dynamically
			csvHeadersLabels[i] = newHeader
		} else {
			csvHeadersLabels[i] = fieldInfo.getFirstKey()
		}
	}
	if !xw.OmitHeaders {
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	inLen := inValue.Len()
	for i := 0; i < inLen; i++ { // Iterate over container rows
		inValueByIndex := inValue.Index(i)
		if xw.OnRecord != nil {
			inValueByIndex = reflect.ValueOf(xw.OnRecord(inValue.Index(i).Interface().(T)))
		}
		for j, fieldInfo := range inInnerStructInfo.Fields {
			csvHeadersLabels[j] = ""
			inInnerFieldValue, err := getInnerField(inValueByIndex, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			csvHeadersLabels[j] = inInnerFieldValue
		}
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	xw.writer.Flush()
	return xw.writer.Error()
}

func (xw *XsvWriter[T]) WriteFromChan(dataChan chan T) error {
	// Get the first value. It wil determine the header structure.
	firstValue, ok := <-dataChan
	if !ok {
		return ErrChannelIsClosed
	}
	inValue, inType := getConcreteReflectValueAndType(firstValue) // Get the concrete type
	if err := ensureStructOrPtr(inType); err != nil {
		return err
	}
	inInnerWasPointer := inType.Kind() == reflect.Ptr
	fieldInfos := getFieldInfos(inType, []int{}, []string{}, xw.TagName, xw.TagSeparator, xw.nameNormalizer) // Get the inner struct info to get CSV annotations
	fieldInfos = xw.getSelectedFieldInfos(fieldInfos)
	if err := xw.checkSortOrderSlice(len(fieldInfos)); err != nil {
		return err
	}
	fieldInfos = reorderColumns[fieldInfo](fieldInfos, xw.SortOrder)
	inInnerStructInfo := &structInfo{fieldInfos}
	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to Write the header (first line) in CSV
		if newHeader, ok := xw.HeaderModifier[fieldInfo.getFirstKey()]; ok { // modify header name dynamically
			csvHeadersLabels[i] = newHeader
		} else {
			csvHeadersLabels[i] = fieldInfo.getFirstKey()
		}
	}

	if !xw.OmitHeaders {
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	write := func(val reflect.Value) error {
		if xw.OnRecord != nil {
			val = reflect.ValueOf(xw.OnRecord(val.Interface().(T)))
		}
		for j, fieldInfo := range inInnerStructInfo.Fields {
			csvHeadersLabels[j] = ""
			inInnerFieldValue, err := getInnerField(val, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			csvHeadersLabels[j] = inInnerFieldValue
		}
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
			return err
		}
		return nil
	}
	if err := write(inValue); err != nil {
		return err
	}
	for v := range dataChan {
		val, _ := getConcreteReflectValueAndType(v) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
		if err := ensureStructOrPtr(inType); err != nil {
			return err
		}
		if err := write(val); err != nil {
			return err
		}
	}
	xw.writer.Flush()
	return xw.writer.Error()
}

func reorderColumns[T any](row []T, sortOrder []int) []T {
	if len(sortOrder) > 1 {
		newLine := make([]T, len(row))
		for from, to := range sortOrder {
			newLine[to] = row[from]
		}
		return newLine
	} else {
		return row
	}
}
