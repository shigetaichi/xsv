// Copyright 2023 Taichi Shigematsu. All rights reserved.
// Use of this source code is governed by a MIT license
// The license can be found in the LICENSE file.

package xsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// FailIfUnmatchedStructTags indicates whether it is considered an error when there is an unmatched
// struct tag.
var FailIfUnmatchedStructTags = false

// FailIfDoubleHeaderNames indicates whether it is considered an error when a header name is repeated
// in the csv header.
var FailIfDoubleHeaderNames = false

// ShouldAlignDuplicateHeadersWithStructFieldOrder indicates whether we should align duplicate CSV
// headers per their alignment in the struct definition.
var ShouldAlignDuplicateHeadersWithStructFieldOrder = false

// TagName defines key in the struct field's tag to scan
var TagName = "csv"

// TagSeparator defines seperator string for multiple csv tags in struct fields
var TagSeparator = ","

// Normalizer is a function that takes and returns a string. It is applied to
// struct and header field values before they are compared. It can be used to alter
// names for comparison. For instance, you could allow case insensitive matching
// or convert '-' to '_'.
type Normalizer func(string) string

type ErrorHandler func(*csv.ParseError) bool

// normalizeName function initially set to a nop Normalizer.
var normalizeName = DefaultNameNormalizer()

// DefaultNameNormalizer is a nop Normalizer.
func DefaultNameNormalizer() Normalizer { return func(s string) string { return s } }

// --------------------------------------------------------------------------
// CSVReader used to parse CSV

var selfCSVReader = DefaultCSVReader

// DefaultCSVReader is the default CSV reader used to parse CSV (cf. csv.NewReader)
func DefaultCSVReader(in io.Reader) CSVReader {
	return csv.NewReader(in)
}

// LazyCSVReader returns a lazy CSV reader, with LazyQuotes and TrimLeadingSpace.
func LazyCSVReader(in io.Reader) CSVReader {
	csvReader := csv.NewReader(in)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	return csvReader
}

func getCSVReader(in io.Reader) CSVReader {
	return selfCSVReader(in)
}

// --------------------------------------------------------------------------
// Marshal functions

// --------------------------------------------------------------------------
// Unmarshal functions

// UnmarshalCSVToMap parses a CSV of 2 columns into a map.
func UnmarshalCSVToMap(in CSVReader, out interface{}) error {
	decoder := NewSimpleDecoderFromCSVReader(in)
	header, err := decoder.GetCSVRow()
	if err != nil {
		return err
	}
	if len(header) != 2 {
		return fmt.Errorf("maps can only be created for csv of two columns")
	}
	outValue, outType := getConcreteReflectValueAndType(out)
	if outType.Kind() != reflect.Map {
		return fmt.Errorf("cannot use " + outType.String() + ", only map supported")
	}
	keyType := outType.Key()
	valueType := outType.Elem()
	outValue.Set(reflect.MakeMap(outType))
	for {
		key := reflect.New(keyType)
		value := reflect.New(valueType)
		line, err := decoder.GetCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := setField(key, line[0], false); err != nil {
			return err
		}
		if err := setField(value, line[1], false); err != nil {
			return err
		}
		outValue.SetMapIndex(key.Elem(), value.Elem())
	}
	return nil
}

// UnmarshalToChan parses the CSV from the reader and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalToChan(in io.Reader, c interface{}) error {
	if c == nil {
		return fmt.Errorf("goscv: channel is %v", c)
	}
	return readEach(csvDecoder{csv.NewReader(in)}, c)
}

// UnmarshalDecoderToChan parses the CSV from the decoder and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalDecoderToChan(in SimpleDecoder, c interface{}) error {
	if c == nil {
		return fmt.Errorf("goscv: channel is %v", c)
	}
	return readEach(in, c)
}

// UnmarshalToCallback parses the CSV from the reader and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalToCallback(in io.Reader, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalToChan(in, c.Interface())
	}()
	for {
		select {
		case err := <-cerr:
			return err
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			break
		}
		callResults := valueFunc.Call([]reflect.Value{v})
		// if last returned value from Call() is an error, return it
		if len(callResults) > 0 {
			if err, ok := callResults[len(callResults)-1].Interface().(error); ok {
				return err
			}
		}
	}
	return nil
}

// UnmarshalDecoderToCallback parses the CSV from the decoder and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalDecoderToCallback(in SimpleDecoder, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalDecoderToChan(in, c.Interface())
	}()
	for {
		select {
		case err := <-cerr:
			return err
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			break
		}
		valueFunc.Call([]reflect.Value{v})
	}
	return nil
}

// UnmarshalBytesToCallback parses the CSV from the bytes and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalBytesToCallback(in []byte, f interface{}) error {
	return UnmarshalToCallback(bytes.NewReader(in), f)
}

// UnmarshalStringToCallback parses the CSV from the string and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalStringToCallback(in string, c interface{}) (err error) {
	return UnmarshalToCallback(strings.NewReader(in), c)
}

// UnmarshalToCallbackWithError parses the CSV from the reader and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalToCallbackWithError(in io.Reader, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	if t.NumOut() != 1 {
		return fmt.Errorf("the given function must have exactly one return value")
	}
	if !isErrorType(t.Out(0)) {
		return fmt.Errorf("the given function must only return error")
	}

	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalToChan(in, c.Interface())
	}()

	var fErr error
	for {
		select {
		case err := <-cerr:
			if err != nil {
				return err
			}
			return fErr
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			if err := <-cerr; err != nil {
				fErr = err
			}
			break
		}

		// callback f has already returned an error, stop processing but keep draining the chan c
		if fErr != nil {
			continue
		}

		results := valueFunc.Call([]reflect.Value{v})

		// If the callback f returns an error, stores it and returns it in future.
		errValue := results[0].Interface()
		if errValue != nil {
			fErr = errValue.(error)
		}
	}
	return fErr
}

// UnmarshalBytesToCallbackWithError parses the CSV from the bytes and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalBytesToCallbackWithError(in []byte, f interface{}) error {
	return UnmarshalToCallbackWithError(bytes.NewReader(in), f)
}

// UnmarshalStringToCallbackWithError parses the CSV from the string and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalStringToCallbackWithError(in string, c interface{}) (err error) {
	return UnmarshalToCallbackWithError(strings.NewReader(in), c)
}

// CSVToMap creates a simple map from a CSV of 2 columns.
func CSVToMap(in io.Reader) (map[string]string, error) {
	decoder := csvDecoder{getCSVReader(in)}
	header, err := decoder.GetCSVRow()
	if err != nil {
		return nil, err
	}
	if len(header) != 2 {
		return nil, fmt.Errorf("maps can only be created for csv of two columns")
	}
	m := make(map[string]string)
	for {
		line, err := decoder.GetCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		m[line[0]] = line[1]
	}
	return m, nil
}

// CSVToMaps takes a reader and returns an array of dictionaries, using the header row as the keys
func CSVToMaps(reader io.Reader) ([]map[string]string, error) {
	r := csv.NewReader(reader)
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows, nil
}

// CSVToChanMaps parses the CSV from the reader and send a dictionary in the chan c, using the header row as the keys.
func CSVToChanMaps(reader io.Reader, c chan<- map[string]string) error {
	r := csv.NewReader(reader)
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			c <- dict
		}
	}
	return nil
}
