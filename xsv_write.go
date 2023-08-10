package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
)

type XsvWrite[T any] struct {
	TagName             string
	TagSeparator        string
	OmitHeaders         bool
	SelectedColumnIndex []int
	ColumnSorter        ColumnSorter
}

func NewXSVWrite[T any]() XsvWrite[T] {
	return XsvWrite[T]{
		TagName:             "csv",
		TagSeparator:        ",",
		OmitHeaders:         false,
		SelectedColumnIndex: make([]int, 0),
		ColumnSorter: func(row []string) []string {
			return row
		},
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
