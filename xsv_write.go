package xsv

import (
	"encoding/csv"
	"os"
)

type XSVWrite[T any] struct {
	Data                StructSlice[T]
	tagName             string
	tagSeparator        string
	fieldsCombiner      string
	selectedColumnIndex []int
	ColumnSorter        ColumnSorter
	OmitHeaders         bool
}
type ColumnSorter = func(row []string) []string

func NewXSVWrite[T any](
	data StructSlice[T],
) XSVWrite[T] {
	return XSVWrite[T]{
		Data:           data,
		tagName:        "csv",
		tagSeparator:   ",",
		fieldsCombiner: ".",
		ColumnSorter: func(row []string) []string {
			return row
		},
	}
}

func (x *XSVWrite[T]) GetTagName() string {
	return x.tagName
}
func (x *XSVWrite[T]) SetTagName(tagName string) {
	x.tagName = tagName
}

func (x *XSVWrite[T]) GetTagSeparator() string {
	return x.tagSeparator
}
func (x *XSVWrite[T]) SetTagSeparator(tagSeparator string) {
	x.tagSeparator = tagSeparator
}

func (x *XSVWrite[T]) GetFieldsCombiner() string {
	return x.fieldsCombiner
}
func (x *XSVWrite[T]) SetFieldsCombiner(fieldsCombiner string) {
	x.fieldsCombiner = fieldsCombiner
}

func (x *XSVWrite[T]) GetColumnIndex() []int {
	return x.selectedColumnIndex
}
func (x *XSVWrite[T]) SetColumnIndex(colIndex []int) {
	x.selectedColumnIndex = colIndex
}

func (x *XSVWrite[T]) Write(writer *csv.Writer) error {
	//TODO: colIndex = changeToSequence(colIndex) // ColumnSorterで各自の実装に任せるってことになりそう
	inValue, inType := getConcreteReflectValueAndType(x.Data) // Get the concrete type (not pointer) (Slice<?> or Array<?>)

	inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureInInnerType(inInnerType); err != nil {
		return err
	}

	inInnerStructInfo := getStructInfoNoCache(inInnerType) // Get the inner struct info to get CSV annotations
	//inInnerStructInfo.Fields = getFilteredFields(inInnerStructInfo.Fields, removeFieldsIndexes) // Filtered out ignoreFields from all fields
	inInnerStructInfo.Fields = getPickedFields(inInnerStructInfo.Fields, x.selectedColumnIndex) // Filter Fields from all fields

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
