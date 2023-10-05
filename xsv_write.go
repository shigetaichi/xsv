package xsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
)

// XsvWrite manages configuration values related to the csv write process.
type XsvWrite[T any] struct {
	TagName         string //key in the struct field's tag to scan
	TagSeparator    string //separator string for multiple csv tags in struct fields
	OmitHeaders     bool
	SelectedColumns []string          // slice of field names to output
	SortOrder       []int             // column sort order
	HeaderModifier  map[string]string // map to dynamically change headers
	OnRecord        func(T, *int) T   // callback function to be called on each record, the int is the index of the record, but i can be nil when executing the writeFromChan function
	nameNormalizer  Normalizer
}

// NewXsvWrite creates a new XsvWrite struct with default configuration values
func NewXsvWrite[T any]() XsvWrite[T] {
	return XsvWrite[T]{
		TagName:         "csv",
		TagSeparator:    ",",
		OmitHeaders:     false,
		SelectedColumns: make([]string, 0),
		SortOrder:       make([]int, 0),
		HeaderModifier:  map[string]string{},
		OnRecord:        nil,
		nameNormalizer:  func(s string) string { return s },
	}
}

func (x *XsvWrite[T]) checkSortOrderSlice(outputFieldsCount int) error {
	if len(x.SortOrder) > 0 {
		if len(x.SortOrder) != outputFieldsCount {
			return errors.New(fmt.Sprintf("the length of the SortOrder array should be equal to the number of items to be output(%d)", outputFieldsCount))
		}
	}
	return nil
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
