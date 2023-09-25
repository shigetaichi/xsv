package xsv

import (
	"encoding/csv"
	"io"
	"reflect"
)

type XsvReader[T any] struct {
	XsvRead[T]
	reader *csv.Reader
}

func NewXsvReader[T any](xsvRead XsvRead[T]) *XsvReader[T] {
	return &XsvReader[T]{XsvRead: xsvRead}
}

func (r *XsvReader[T]) Lazy() *XsvReader[T] {
	r.reader.LazyQuotes = true
	r.reader.TrimLeadingSpace = true
	return r
}

func (r *XsvReader[T]) ReadTo(out *[]T) error {
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)

	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := r.reader.ReadAll() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if len(csvRows) == 0 {
		return ErrEmptyCSVFile
	}

	if err := r.checkFromTo(); err != nil {
		return err
	}
	to := r.To
	if to >= 0 {
		to++
	} else {
		to = len(csvRows)
	}
	body := csvRows[r.From:to]
	capacity := len(body) + 1                                      // Plus one for the header row.
	if err := ensureOutCapacity(&outValue, capacity); err != nil { // Ensure the container is big enough to hold the CSV content
		return err
	}
	fieldInfos := getFieldInfos(outInnerType, []int{}, []string{}, r.TagName, r.TagSeparator, r.NameNormalizer) // Get the inner struct info to get CSV annotations
	outInnerStructInfo := &structInfo{fieldInfos}
	if len(outInnerStructInfo.Fields) == 0 {
		return ErrNoStructTags
	}

	headers := make([]string, len(csvRows[0]))
	for i, h := range csvRows[0] { // apply normalizer func to headers
		headers[i] = r.NameNormalizer(h)
	}

	csvHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV

	headerCount := map[string]int{}
	for i, csvColumnHeader := range headers {
		curHeaderCount := headerCount[csvColumnHeader]
		if fieldInfo := getCSVFieldPosition(csvColumnHeader, outInnerStructInfo, curHeaderCount); fieldInfo != nil {
			csvHeadersLabels[i] = fieldInfo
			if r.ShouldAlignDuplicateHeadersWithStructFieldOrder {
				curHeaderCount++
				headerCount[csvColumnHeader] = curHeaderCount
			}
		}
	}

	if r.FailIfUnmatchedStructTags {
		if err := maybeMissingStructFields(outInnerStructInfo.Fields, headers); err != nil {
			return err
		}
	}
	if r.FailIfDoubleHeaderNames {
		if err := maybeDoubleHeaderNames(headers); err != nil {
			return err
		}
	}

	var withFieldsOK bool
	var fieldTypeUnmarshallerWithKeys TypeUnmarshalCSVWithFields

	for i, csvRow := range body {
		objectIface := reflect.New(outValue.Index(i).Type()).Interface()
		outInner := createNewOutInner(outInnerWasPointer, outInnerType)
		for j, csvColumnContent := range csvRow {
			if fieldInfo, ok := csvHeadersLabels[j]; ok { // Position found accordingly to header name

				if outInner.CanInterface() {
					fieldTypeUnmarshallerWithKeys, withFieldsOK = objectIface.(TypeUnmarshalCSVWithFields)
					if withFieldsOK {
						if err := fieldTypeUnmarshallerWithKeys.UnmarshalCSVWithFields(fieldInfo.getFirstKey(), csvColumnContent); err != nil {
							parseError := csv.ParseError{
								Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
								Column: j + 1,
								Err:    err,
							}
							return &parseError
						}
						continue
					}
				}
				value := csvColumnContent
				if value == "" {
					value = fieldInfo.defaultValue
				}
				if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, value, fieldInfo.omitEmpty); err != nil { // Set field of struct
					parseError := csv.ParseError{
						Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
						Column: j + 1,
						Err:    err,
					}
					if r.ErrorHandler == nil || !r.ErrorHandler(&parseError) {
						return &parseError
					}
				}
			}
		}

		if withFieldsOK {
			reflectedObject := reflect.ValueOf(objectIface)
			outInner = reflectedObject.Elem()
		}

		if r.OnRecord != nil {
			outInner = reflect.ValueOf(r.OnRecord(outInner.Interface().(T)))
		}

		outValue.Index(i).Set(outInner)
	}
	return nil
}

func (r *XsvReader[T]) ReadEach(c chan T) error {
	outValue, outType := getConcreteReflectValueAndType(c) // Get the concrete type (not pointer)
	defer close(c)

	headers, err := r.reader.Read()
	if err != nil {
		return err
	}

	for i, h := range headers { // apply normalizer func to headers
		headers[i] = r.NameNormalizer(h)
	}

	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	fieldInfos := getFieldInfos(outInnerType, []int{}, []string{}, r.TagName, r.TagSeparator, r.NameNormalizer) // Get the inner struct info to get CSV annotations
	outInnerStructInfo := &structInfo{fieldInfos}
	if len(outInnerStructInfo.Fields) == 0 {
		return ErrNoStructTags
	}
	csvHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV
	headerCount := map[string]int{}
	for i, csvColumnHeader := range headers {
		curHeaderCount := headerCount[csvColumnHeader]
		if fieldInfo := getCSVFieldPosition(csvColumnHeader, outInnerStructInfo, curHeaderCount); fieldInfo != nil {
			csvHeadersLabels[i] = fieldInfo
			if r.ShouldAlignDuplicateHeadersWithStructFieldOrder {
				curHeaderCount++
				headerCount[csvColumnHeader] = curHeaderCount
			}
		}
	}
	if err := maybeMissingStructFields(outInnerStructInfo.Fields, headers); err != nil {
		if r.FailIfUnmatchedStructTags {
			return err
		}
	}
	if r.FailIfDoubleHeaderNames {
		if err := maybeDoubleHeaderNames(headers); err != nil {
			return err
		}
	}

	if err := r.checkFromTo(); err != nil {
		return err
	}
	i := 1
	for {
		line, err := r.reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if r.From <= i && i <= r.To {
			outInner := createNewOutInner(outInnerWasPointer, outInnerType)
			for j, csvColumnContent := range line {
				if fieldInfo, ok := csvHeadersLabels[j]; ok { // Position found accordingly to header name
					if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
						return &csv.ParseError{
							Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
							Column: j + 1,
							Err:    err,
						}
					}
				}
			}
			if r.OnRecord != nil {
				outInner = reflect.ValueOf(r.OnRecord(outInner.Interface().(T)))
			}
			outValue.Send(outInner)
		}
		i++
	}
	return nil
}

func (r *XsvReader[T]) ReadToWithoutHeaders(out *[]T) error {
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)

	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := r.reader.ReadAll() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if len(csvRows) == 0 {
		return ErrEmptyCSVFile
	}

	if err := r.checkFromTo(); err != nil {
		return err
	}
	to := r.To
	if to >= 0 {
		to++
	} else {
		to = len(csvRows)
	}
	body := csvRows[r.From:to]
	capacity := len(body) + 1
	if err := ensureOutCapacity(&outValue, capacity); err != nil { // Ensure the container is big enough to hold the CSV content
		return err
	}
	fieldInfos := getFieldInfos(outInnerType, []int{}, []string{}, r.TagName, r.TagSeparator, r.NameNormalizer) // Get the inner struct info to get CSV annotations
	outInnerStructInfo := &structInfo{fieldInfos}
	if len(outInnerStructInfo.Fields) == 0 {
		return ErrNoStructTags
	}

	for i, csvRow := range csvRows {
		outInner := createNewOutInner(outInnerWasPointer, outInnerType)
		for j, csvColumnContent := range csvRow {
			fieldInfo := outInnerStructInfo.Fields[j]
			if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
				return &csv.ParseError{
					Line:   i + 1,
					Column: j + 1,
					Err:    err,
				}
			}
		}
		if r.OnRecord != nil {
			outInner = reflect.ValueOf(r.OnRecord(outInner.Interface().(T)))
		}
		outValue.Index(i).Set(outInner)
	}

	return nil
}

func (r *XsvReader[T]) ReadEachWithoutHeaders(c chan T) error {
	outValue, outType := getConcreteReflectValueAndType(c) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	defer close(c)

	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	fieldInfos := getFieldInfos(outInnerType, []int{}, []string{}, r.TagName, r.TagSeparator, r.NameNormalizer) // Get the inner struct info to get CSV annotations
	outInnerStructInfo := &structInfo{fieldInfos}
	if len(outInnerStructInfo.Fields) == 0 {
		return ErrNoStructTags
	}

	if err := r.checkFromTo(); err != nil {
		return err
	}
	i := 1
	for {
		line, err := r.reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if r.From <= i && i <= r.To {
			outInner := createNewOutInner(outInnerWasPointer, outInnerType)
			for j, csvColumnContent := range line {
				fieldInfo := outInnerStructInfo.Fields[j]
				if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
					return &csv.ParseError{
						Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
						Column: j + 1,
						Err:    err,
					}
				}
			}
			if r.OnRecord != nil {
				outInner = reflect.ValueOf(r.OnRecord(outInner.Interface().(T)))
			}
			outValue.Send(outInner)
		}
		i++
	}
	return nil
}

func (r *XsvReader[T]) ReadToCallback(f func(s T) error) error {
	cerr := make(chan error)
	c := make(chan T)
	go func() {
		cerr <- r.ReadEach(c)
	}()
	for {
		select {
		case err := <-cerr:
			return err
		case v, ok := <-c:
			if !ok {
				break
			}
			if err := f(v); err != nil {
				return err
			}
		default:
		}
	}
}

func (r *XsvReader[T]) ToMap() ([]map[string]string, error) {
	var rows []map[string]string
	var header []string
	var i = 0

	if err := r.checkFromTo(); err != nil {
		return nil, err
	}
	for {
		record, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header == nil {
			header = record
		} else {
			if r.From <= i && i <= r.To {
				dict := map[string]string{}
				for i := range header {
					dict[header[i]] = record[i]
				}
				if r.OnRecord != nil {
					v := r.OnRecord(reflect.ValueOf(dict).Interface().(T))
					dict = reflect.ValueOf(v).Interface().(map[string]string)
				}
				rows = append(rows, dict)
			}
		}
		i++
	}
	return rows, nil
}

func (r *XsvReader[T]) ToChanMaps(c chan<- map[string]string) error {
	var header []string
	var i = 0

	if err := r.checkFromTo(); err != nil {
		return err
	}
	for {
		record, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header == nil {
			header = record
		} else {
			if r.From <= i && i <= r.To {
				dict := map[string]string{}
				for i := range header {
					dict[header[i]] = record[i]
				}
				if r.OnRecord != nil {
					v := r.OnRecord(reflect.ValueOf(dict).Interface().(T))
					dict = reflect.ValueOf(v).Interface().(map[string]string)
				}
				c <- dict
			}
		}
		i++
	}
	return nil
}
