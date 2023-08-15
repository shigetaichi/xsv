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
	inInnerStructInfo := &structInfo{fieldInfos}

	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}
	if err := xw.checkSortOrderSlice(len(fieldInfos)); err != nil {
		return err
	}
	csvHeadersLabels = reorderColumns(csvHeadersLabels, xw.SortOrder)
	if !xw.OmitHeaders {
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
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
		csvHeadersLabels = reorderColumns(csvHeadersLabels, xw.SortOrder)
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
	inInnerStructInfo := &structInfo{fieldInfos}
	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to Write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}

	if err := xw.checkSortOrderSlice(len(fieldInfos)); err != nil {
		return err
	}
	csvHeadersLabels = reorderColumns(csvHeadersLabels, xw.SortOrder)
	if !xw.OmitHeaders {
		if err := xw.writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	write := func(val reflect.Value) error {
		for j, fieldInfo := range inInnerStructInfo.Fields {
			csvHeadersLabels[j] = ""
			inInnerFieldValue, err := getInnerField(val, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			csvHeadersLabels[j] = inInnerFieldValue
			csvHeadersLabels = reorderColumns(csvHeadersLabels, xw.SortOrder)
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

func reorderColumns(row []string, sortOrder []int) []string {
	if len(sortOrder) > 0 {
		newLine := make([]string, len(row))
		for from, to := range sortOrder {
			newLine[to] = row[from]
		}
		return newLine
	} else {
		return row
	}
}
