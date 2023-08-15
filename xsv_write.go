package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
	"slices"
)

type XsvWrite[T any] struct {
	TagName         string //key in the struct field's tag to scan
	TagSeparator    string //separator string for multiple csv tags in struct fields
	OmitHeaders     bool
	SelectedColumns []string     // slice of field names to output
	columnSorter    ColumnSorter // TODO: describe in comment
	nameNormalizer  Normalizer
}
type ColumnSorter = func(row []string) []string

func NewXsvWrite[T any]() XsvWrite[T] {
	return XsvWrite[T]{
		TagName:         "csv",
		TagSeparator:    ",",
		OmitHeaders:     false,
		SelectedColumns: make([]string, 0),
		columnSorter: func(row []string) []string {
			return row
		},
		nameNormalizer: func(s string) string { return s },
	}
}

func (x *XsvWrite[T]) getSelectedFieldInfos(fieldInfos []fieldInfo) []fieldInfo {
	if len(x.SelectedColumns) > 0 {
		var selectedFieldInfos []fieldInfo
		for _, info := range fieldInfos {
			if slices.Index(x.SelectedColumns, info.keys[0]) >= 0 {
				selectedFieldInfos = append(selectedFieldInfos, info)
			}
		}
		return selectedFieldInfos
	} else {
		return fieldInfos
	}
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
