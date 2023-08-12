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
	decoder := csvDecoder{r.reader}
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := ensureOutType(outType); err != nil {
		return err
	}
	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := decoder.GetCSVRows() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if len(csvRows) == 0 {
		return ErrEmptyCSVFile
	}
	if err := ensureOutCapacity(&outValue, len(csvRows)); err != nil { // Ensure the container is big enough to hold the CSV content
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
	body := csvRows[1:]

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

		outValue.Index(i).Set(outInner)
	}
	return nil
}

func (r *XsvReader[T]) ReadEach(c chan T) error {
	decoder := csvDecoder{r.reader}
	outValue, outType := getConcreteReflectValueAndType(c) // Get the concrete type (not pointer)
	//if outType.Kind() != reflect.Chan {                    // TODO: 不要な場合は削除
	//	return fmt.Errorf("cannot use %v with type %s, only channel supported", c, outType)
	//}
	//defer outValue.Close()
	defer close(c)

	headers, err := decoder.GetCSVRow()
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
	i := 0
	for {
		line, err := decoder.GetCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
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
		outValue.Send(outInner)
		i++
	}
	return nil
}

func (r *XsvReader[T]) ReadToWithoutHeaders(out *[]T) error {
	decoder := csvDecoder{r.reader}
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := ensureOutType(outType); err != nil {
		return err
	}
	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := decoder.GetCSVRows() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if len(csvRows) == 0 {
		return ErrEmptyCSVFile
	}
	if err := ensureOutCapacity(&outValue, len(csvRows)+1); err != nil { // Ensure the container is big enough to hold the CSV content
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
		outValue.Index(i).Set(outInner)
	}

	return nil
}

func (r *XsvReader[T]) ReadEachWithoutHeaders(c chan T) error {
	decoder := csvDecoder{r.reader}
	outValue, outType := getConcreteReflectValueAndType(c) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	//if err := ensureOutType(outType); err != nil { // TODO: 不要なら消す
	//	return err
	//}
	//defer outValue.Close()
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

	i := 0
	for {
		line, err := decoder.GetCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
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
		outValue.Send(outInner)
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
