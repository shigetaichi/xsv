// Copyright 2023 Taichi Shigematsu. All rights reserved.
// Use of this source code is governed by a MIT license
// The license can be found in the LICENSE file.

package xsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
)

// --------------------------------------------------------------------------
// Unmarshal functions

// UnmarshalCSVToMap parses a CSV of 2 columns into a map.
func UnmarshalCSVToMap(r *csv.Reader, out interface{}) error {
	header, err := r.Read()
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
		line, err := r.Read()
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

// CSVToMap creates a simple map from a CSV of 2 columns.
func CSVToMap(in io.Reader) (map[string]string, error) {
	r := csv.NewReader(in)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(header) != 2 {
		return nil, fmt.Errorf("maps can only be created for csv of two columns")
	}
	m := make(map[string]string)
	for {
		line, err := r.Read()
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
