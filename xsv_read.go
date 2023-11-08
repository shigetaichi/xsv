package xsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// XsvRead manages configuration values related to the csv read process.
type XsvRead[T any] struct {
	TagName                                         string //key in the struct field's tag to scan
	TagSeparator                                    string //separator string for multiple csv tags in struct fields
	FailIfUnmatchedStructTags                       bool   // indicates whether it is considered an error when there is an unmatched struct tag.
	FailIfDoubleHeaderNames                         bool   // indicates whether it is considered an error when a header name is repeated in the csv header.
	ShouldAlignDuplicateHeadersWithStructFieldOrder bool   // indicates whether we should align duplicate CSV headers per their alignment in the struct definition.
	From                                            int    //
	To                                              int    //
	OnRecord                                        func(T) T // callback function to be called on each record
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
		From:           1,
		To:             -1,
		OnRecord:       nil,
		NameNormalizer: func(s string) string { return s },
		ErrorHandler:   nil,
	}
}

func (x *XsvRead[T]) checkFrom() (err error) {
	if x.From >= 0 {
		return nil
	}
	return errors.New(fmt.Sprintf("%s cannot be set to a negative value.", strconv.Quote("From")))
}

func (x *XsvRead[T]) checkTo() (err error) {
	if x.To >= -1 {
		return nil
	}
	return errors.New(fmt.Sprintf("%s cannot be set to a negative value other than -1.", strconv.Quote("To")))
}

func (x *XsvRead[T]) checkFromTo() (err error) {
	if err := x.checkFrom(); err != nil {
		return err
	}
	if err := x.checkTo(); err != nil {
		return err
	}
	if x.To == -1 {
		return nil
	}
	if x.From <= x.To {
		return nil
	}
	return errors.New(fmt.Sprintf("%s cannot be set before %s", strconv.Quote("To"), strconv.Quote("From")))
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
