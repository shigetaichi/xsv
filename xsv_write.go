package xsv

import (
	"encoding/csv"
	"os"
	"reflect"
)

type XSVWrite[T any] struct {
	Data                StructSlice[T]
	DataChan            chan T
	TagName             string
	TagSeparator        string
	OmitHeaders         bool
	SelectedColumnIndex []int
	ColumnSorter        ColumnSorter
}
type ColumnSorter = func(row []string) []string

func NewXSVWrite[T any](
	data StructSlice[T],
) XSVWrite[T] {
	return XSVWrite[T]{
		Data:                data,
		DataChan:            make(chan T),
		TagName:             "csv",
		TagSeparator:        ",",
		OmitHeaders:         false,
		SelectedColumnIndex: make([]int, 0),
		ColumnSorter: func(row []string) []string {
			return row
		},
	}
}

func (x *XSVWrite[T]) WriteFromChan(writer *csv.Writer) error {
	// Get the first value. It wil determine the header structure.
	firstValue, ok := <-x.DataChan
	if !ok {
		return ErrChannelIsClosed
	}
	inValue, inType := getConcreteReflectValueAndType(firstValue) // Get the concrete type
	if err := ensureStructOrPtr(inType); err != nil {
		return err
	}
	inInnerWasPointer := inType.Kind() == reflect.Ptr
	inInnerStructInfo := getStructInfoNoCache(inType)                                           // Get the inner struct info to get CSV annotations
	inInnerStructInfo.Fields = getPickedFields(inInnerStructInfo.Fields, x.SelectedColumnIndex) // Filtered out ignoreFields from all fields
	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}
	if !x.OmitHeaders {
		if err := writer.Write(csvHeadersLabels); err != nil {
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
			csvHeadersLabels = x.ColumnSorter(csvHeadersLabels)
		}
		if err := writer.Write(csvHeadersLabels); err != nil {
			return err
		}
		return nil
	}
	if err := write(inValue); err != nil {
		return err
	}
	for v := range x.DataChan {
		val, _ := getConcreteReflectValueAndType(v) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
		if err := ensureStructOrPtr(inType); err != nil {
			return err
		}
		if err := write(val); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func (x *XSVWrite[T]) WriteFromChanToFile(file *os.File) error {
	w := getCSVWriter(file)
	return x.WriteFromChan(w.Writer)
}

func (x *XSVWrite[T]) Write(writer *csv.Writer) error {
	inValue, inType := getConcreteReflectValueAndType(x.Data) // Get the concrete type (not pointer) (Slice<?> or Array<?>)

	inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureInInnerType(inInnerType); err != nil {
		return err
	}

	fieldsList := getFieldInfosWithTagName(inInnerType, []int{}, []string{}, x.TagName, x.TagSeparator) // Get the inner struct info to get CSV annotations
	inInnerStructInfo := &structInfo{fieldsList}

	inInnerStructInfo.Fields = getPickedFields(inInnerStructInfo.Fields, x.SelectedColumnIndex) // Filter Fields from all fields

	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}
	csvHeadersLabels = x.ColumnSorter(csvHeadersLabels)
	if !x.OmitHeaders {
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
		csvHeadersLabels = x.ColumnSorter(csvHeadersLabels)
		if err := writer.Write(csvHeadersLabels); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func (x *XSVWrite[T]) WriteToFile(file *os.File) error {
	w := getCSVWriter(file)
	return x.Write(w.Writer)
}
