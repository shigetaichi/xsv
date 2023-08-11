package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"
)

type XsvRead[T any] struct {
	TagName                                         string
	TagSeparator                                    string
	FailIfUnmatchedStructTags                       bool
	FailIfDoubleHeaderNames                         bool
	ShouldAlignDuplicateHeadersWithStructFieldOrder bool
	NameNormalizer                                  Normalizer
	ErrorHandler                                    ErrorHandler
}

func NewXSVRead[T any]() *XsvRead[T] {
	return &XsvRead[T]{
		TagName:                   "csv",
		TagSeparator:              ",",
		FailIfUnmatchedStructTags: false,
		FailIfDoubleHeaderNames:   false,
		ShouldAlignDuplicateHeadersWithStructFieldOrder: false,
		NameNormalizer: func(s string) string { return s },
		ErrorHandler:   nil,
	}
}

func (x *XsvRead[T]) SetReader(r *csv.Reader) (xr *XsvReader[T]) {
	xr = NewXsvReader(*x)
	xr.reader = r
	return xr
}

func (x *XsvRead[T]) SetFileReader(file *os.File) (xr *XsvReader[T]) {
	return x.SetReader(csv.NewReader(file))
}

func (x *XsvRead[T]) SetStringReader(string string) (xr *XsvReader[T]) {
	return x.SetReader(csv.NewReader(strings.NewReader(string)))
}

func (x *XsvRead[T]) SetByteReader(byte []byte) (xr *XsvReader[T]) {
	return x.SetReader(csv.NewReader(bytes.NewReader(byte)))
}
