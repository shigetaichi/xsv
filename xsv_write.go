package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
)

type XsvWrite[T any] struct {
	TagName             string //key in the struct field's tag to scan
	TagSeparator        string //separator string for multiple csv tags in struct fields
	OmitHeaders         bool
	selectedColumnIndex []int        // TODO: describe in comment
	columnSorter        ColumnSorter // TODO: describe in comment
	nameNormalizer      Normalizer
}

func NewXsvWrite[T any]() XsvWrite[T] {
	return XsvWrite[T]{
		TagName:             "csv",
		TagSeparator:        ",",
		OmitHeaders:         false,
		selectedColumnIndex: make([]int, 0),
		columnSorter: func(row []string) []string {
			return row
		},
		nameNormalizer: func(s string) string { return s },
	}
}

type ColumnSorter = func(row []string) []string

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
