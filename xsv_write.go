package xsv

import (
	"encoding/csv"
	"os"
)

type XSVWrite[T any] struct {
	Data                StructSlice[T]
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
		TagName:             "csv",
		TagSeparator:        ",",
		OmitHeaders:         false,
		SelectedColumnIndex: make([]int, 0),
		ColumnSorter: func(row []string) []string {
			return row
		},
	}
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
