package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
	"reflect"
	"slices"
)

type XsvWrite[T any] struct {
	TagName             string //key in the struct field's tag to scan
	TagSeparator        string //separator string for multiple csv tags in struct fields
	OmitHeaders         bool
	selectedColumns []string        // slice indexes of selected columns
	columnSorter        ColumnSorter // TODO: describe in comment
	nameNormalizer      Normalizer
}
type ColumnSorter = func(row []string) []string

func NewXsvWrite[T any]() XsvWrite[T] {
	return XsvWrite[T]{
		TagName:             "csv",
		TagSeparator:        ",",
		OmitHeaders:         false,
		selectedColumns: make([]string, 0),
		columnSorter: func(row []string) []string {
			return row
		},
		nameNormalizer: func(s string) string { return s },
	}
}

func (x *XsvWrite[T]) getIndexesOfSelectedColumns() (columnFieldsIndexes []int) {
	var writeDataType T
	field := reflect.TypeOf(writeDataType)
	var fieldNames []string
	for i := 0; i < field.NumField(); i++ { // TODO:もっといいやり方あるはず。
		fieldNames = append(fieldNames, field.Field(i).Tag.Get(x.TagName))
	}
	for _, column := range x.selectedColumns {
		columnFieldsIndexes = append(columnFieldsIndexes, slices.Index(fieldNames, column))
	}
	return columnFieldsIndexes
}

func (x *XsvWrite[T]) SetWriter(writer *csv.Writer) (xw *XsvWriter[T]) {
	xw = NewXsvWriter(*x)
	xw.writer = writer
	return xw
}

func (x *XsvWrite[T]) SetFileWriter(file *os.File) (xw *XsvWriter[T]) {
	xw = x.SetWriter(csv.NewWriter(file))
	return xw
}

func (x *XsvWrite[T]) SetBufferWriter(buffer *bytes.Buffer) (xw *XsvWriter[T]) {
	xw = x.SetWriter(csv.NewWriter(buffer))
	return xw
}
