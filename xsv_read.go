package xsv

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"
)

// XsvRead manages configuration values related to the csv read process.
type XsvRead[T any] struct {
	TagName                                         string //key in the struct field's tag to scan
	TagSeparator                                    string //separator string for multiple csv tags in struct fields
	FailIfUnmatchedStructTags                       bool   // indicates whether it is considered an error when there is an unmatched struct tag.
	FailIfDoubleHeaderNames                         bool   // indicates whether it is considered an error when a header name is repeated in the csv header.
	ShouldAlignDuplicateHeadersWithStructFieldOrder bool   // indicates whether we should align duplicate CSV headers per their alignment in the struct definition.
	NameNormalizer                                  Normalizer
	ErrorHandler                                    ErrorHandler
}

// NewXsvRead creates a new XsvRead struct with default configuration values
func NewXsvRead[T any]() *XsvRead[T] {
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
