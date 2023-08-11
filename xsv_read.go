package xsv

import (
	"encoding/csv"
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
