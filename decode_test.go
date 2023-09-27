package xsv

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_readTo(t *testing.T) {
	blah := 0
	sptr := "*string"
	sptr2 := ""
	b := bytes.NewBufferString(`foo,BAR,Baz,Blah,SPtr,Omit
f,1,baz,,*string,*string
e,3,b,,,`)
	var samples []Sample
	xsvRead := NewXsvRead[Sample]()
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz", Blah: &blah, SPtr: &sptr, Omit: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}

	expected = Sample{Foo: "e", Bar: 3, Baz: "b", Blah: &blah, SPtr: &sptr2}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}

	b = bytes.NewBufferString(`foo,BAR,Baz
f,1,baz
e,BAD_INPUT,b`)

	samples = []Sample{}
	err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples)
	if err == nil {
		t.Fatalf("Expected error from bad input, got: %+v", samples)
	}
	switch actualErr := err.(type) {
	case *csv.ParseError:
		if actualErr.Line != 3 {
			t.Fatalf("Expected csv.ParseError on line 3, got: %d", actualErr.Line)
		}
		if actualErr.Column != 2 {
			t.Fatalf("Expected csv.ParseError in column 2, got: %d", actualErr.Column)
		}
	default:
		t.Fatalf("incorrect error type: %T", err)
	}

}

func Test_readTo_OnRecord(t *testing.T) {
	blah := 0
	sptr := "*string"
	b := bytes.NewBufferString(`foo,BAR,Baz,Blah,SPtr,Omit
f,1,baz,,*string,*string`)
	var samples []Sample
	xsvRead := NewXsvRead[Sample]()
	xsvRead.OnRecord = func(sample Sample) Sample {
		if sample.Foo == "f" {
			sample.Foo = "f-onrecord"
		}
		return sample
	}
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}

	if len(samples) != 1 {
		t.Fatalf("expected 1 sample instances, got %d", len(samples))
	}
	expected := Sample{Foo: "f-onrecord", Bar: 1, Baz: "baz", Blah: &blah, SPtr: &sptr, Omit: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
}

func Test_readToNormalized(t *testing.T) {

	blah := 0
	sptr := "*string"
	sptr2 := ""
	b := bytes.NewBufferString(`FOO,BAR,BAZ,BLAH,SPTR,OMIT
f,1,baz,,*string,*string
e,3,b,,,`)
	var samples []Sample
	xsvRead := NewXsvRead[Sample]()
	xsvRead.NameNormalizer = func(s string) string {
		return strings.ToLower(s)
	}
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz", Blah: &blah, SPtr: &sptr, Omit: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}

	expected = Sample{Foo: "e", Bar: 3, Baz: "b", Blah: &blah, SPtr: &sptr2}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}

	b = bytes.NewBufferString(`foo,BAR,Baz
f,1,baz
e,BAD_INPUT,b`)

	samples = []Sample{}
	err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples)
	if err == nil {
		t.Fatalf("Expected error from bad input, got: %+v", samples)
	}
	switch actualErr := err.(type) {
	case *csv.ParseError:
		if actualErr.Line != 3 {
			t.Fatalf("Expected csv.ParseError on line 3, got: %d", actualErr.Line)
		}
		if actualErr.Column != 2 {
			t.Fatalf("Expected csv.ParseError in column 2, got: %d", actualErr.Column)
		}
	default:
		t.Fatalf("incorrect error type: %T", err)
	}

}

func Test_readTo_Time(t *testing.T) {
	b := bytes.NewBufferString(`Foo
1970-01-01T03:01:00+03:00`)

	var samples []DateTime
	if err := NewXsvRead[DateTime]().SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}

	rt, _ := time.Parse(time.RFC3339, "1970-01-01T03:01:00+03:00")

	expected := DateTime{Foo: rt}

	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
}

func Test_readTo_complex_embed(t *testing.T) {
	b := bytes.NewBufferString(`first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee
ff,gg,22,hh,ii,jj`)

	var samples []SkipFieldSample
	if err := NewXsvRead[SkipFieldSample]().SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}
	expected := SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "aa",
			Sample: Sample{
				Foo: "bb",
				Bar: 11,
				Baz: "cc",
			},
			Quux: "dd",
		},
		Corge: "ee",
	}
	if expected != samples[0] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
	expected = SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "ff",
			Sample: Sample{
				Foo: "gg",
				Bar: 22,
				Baz: "hh",
			},
			Quux: "ii",
		},
		Corge: "jj",
	}
	if expected != samples[1] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[1])
	}
}

func Test_readTo_embed_ptr(t *testing.T) {
	b := bytes.NewBufferString(`first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee
ff,gg,22,hh,ii,jj`)

	var rows []EmbedPtrSample
	if err := NewXsvRead[EmbedPtrSample]().SetReader(csv.NewReader(b)).ReadTo(&rows); err != nil {
		t.Fatalf(err.Error())
	}
	expected := EmbedPtrSample{
		Qux: "ff",
		Sample: &Sample{
			Foo: "gg",
			Bar: 22,
			Baz: "hh",
		},
		Quux: "ii",
	}
	if !reflect.DeepEqual(expected, rows[1]) {
		t.Fatalf("expected first sample %v, got %+v", expected, rows[1])
	}
}

func Test_readTo_slice(t *testing.T) {
	b := bytes.NewBufferString(`Slice
[]
[1, 2, 3]`)
	reader := csv.NewReader(b)
	reader.Comma = '\t'
	samples := []SliceSample{}
	if err := NewXsvRead[SliceSample]().SetReader(reader).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	expected := SliceSample{Slice: []int{}}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0].Slice)
	}
	expected = SliceSample{Slice: []int{1, 2, 3}}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1].Slice)
	}
}

func Test_readTo_slice_structs(t *testing.T) {
	b := bytes.NewBufferString(`s[0].string,slice[0].f,slice[1].s,s[1].float,a[0].s,array[0].float,a[1].s,array[1].float,ints[0],ints[1],ints[2]
s1,1.1,s2,2.2,s3,3.3,s4,4.4,1,2,3`)

	var samples []SliceStructSample
	err := NewXsvRead[SliceStructSample]().SetReader(csv.NewReader(b)).ReadTo(&samples)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expected := SliceStructSample{
		Slice: []SliceStruct{
			{String: "s1", Float: 1.1},
			{String: "s2", Float: 2.2},
		},
		SimpleSlice: []int{1, 2, 3},
		Array: [2]SliceStruct{
			{String: "s3", Float: 3.3},
			{String: "s4", Float: 4.4},
		},
	}

	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected \n  sample: %v\n     got: %v", expected, samples[0])
	}
}

func Test_readTo_embed_marshal(t *testing.T) {
	b := bytes.NewBufferString(`foo
bar`)

	var rows []EmbedMarshal
	if err := NewXsvRead[EmbedMarshal]().SetReader(csv.NewReader(b)).ReadTo(&rows); err != nil {
		t.Fatalf(err.Error())
	}
	expected := EmbedMarshal{
		Foo: &MarshalSample{Dummy: "bar"},
	}
	if !reflect.DeepEqual(expected, rows[0]) {
		t.Fatalf("expected first sample %v, got %+v", expected, rows[1])
	}
}

func Test_readTo_embed_unmarshal_csv_with_clashing_field(t *testing.T) {
	b := bytes.NewBufferString(`Symbol,Timestamp
test,1656460798.693201614`)

	var rows []EmbedUnmarshalCSVWithClashingField
	if err := NewXsvRead[EmbedUnmarshalCSVWithClashingField]().SetReader(csv.NewReader(b)).ReadTo(&rows); err != nil {
		t.Fatalf(err.Error())
	}
	expected := EmbedUnmarshalCSVWithClashingField{
		Symbol:    "test",
		Timestamp: &UnmarshalCSVSample{Timestamp: 1656460798, Nanos: 693201614},
	}
	if !reflect.DeepEqual(expected, rows[0]) {
		t.Fatalf("expected first sample %v, got %+v", expected, rows[0])
	}
}

func Test_readEach(t *testing.T) {
	b := bytes.NewBufferString(`first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee
ff,gg,22,hh,ii,jj`)

	c := make(chan SkipFieldSample)
	var samples []SkipFieldSample
	go func() {
		if err := NewXsvRead[SkipFieldSample]().SetReader(csv.NewReader(b)).ReadEach(c); err != nil {
			t.Fatal(err)
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}
	expected := SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "aa",
			Sample: Sample{
				Foo: "bb",
				Bar: 11,
				Baz: "cc",
			},
			Quux: "dd",
		},
		Corge: "ee",
	}
	if expected != samples[0] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
	expected = SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "ff",
			Sample: Sample{
				Foo: "gg",
				Bar: 22,
				Baz: "hh",
			},
			Quux: "ii",
		},
		Corge: "jj",
	}
	if expected != samples[1] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[1])
	}
}

func Test_readEach_OnRecord(t *testing.T) {
	b := bytes.NewBufferString(`first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee`)
	c := make(chan SkipFieldSample)
	var samples []SkipFieldSample
	xsvRead := NewXsvRead[SkipFieldSample]()
	go func() {
		xsvRead.OnRecord = func(sample SkipFieldSample) SkipFieldSample {
			if sample.Foo == "bb" {
				sample.Foo = "bb-onrecord"
			}
			return sample
		}
		if err := xsvRead.SetReader(csv.NewReader(b)).ReadEach(c); err != nil {
			t.Fatal(err)
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample instances, got %d", len(samples))
	}
	expected := SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "aa",
			Sample: Sample{
				Foo: "bb-onrecord",
				Bar: 11,
				Baz: "cc",
			},
			Quux: "dd",
		},
		Corge: "ee",
	}
	if expected != samples[0] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
}

func Test_readEach_FromTo(t *testing.T) {
	b := bytes.NewBufferString(`
first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee
ff,gg,22,hh,ii,jj
kk,ll,33,mm,nn,oo
`)

	c := make(chan SkipFieldSample)
	var samples []SkipFieldSample
	go func() {
		xsvRead := NewXsvRead[SkipFieldSample]()
		xsvRead.From = 1
		xsvRead.To = 2
		if err := xsvRead.SetReader(csv.NewReader(b)).ReadEach(c); err != nil {
			t.Fatal(err)
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}
	expected := SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "aa",
			Sample: Sample{
				Foo: "bb",
				Bar: 11,
				Baz: "cc",
			},
			Quux: "dd",
		},
		Corge: "ee",
	}
	if expected != samples[0] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
	expected = SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "ff",
			Sample: Sample{
				Foo: "gg",
				Bar: 22,
				Baz: "hh",
			},
			Quux: "ii",
		},
		Corge: "jj",
	}
	if expected != samples[1] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[1])
	}
}

func Test_readEachWithoutHeaders(t *testing.T) {
	blah := 0
	sptr := ""
	b := bytes.NewBufferString(`f,1,baz,1.66,,,
e,3,b,,,,`)

	c := make(chan Sample)
	var samples []Sample
	go func() {
		if err := NewXsvRead[Sample]().SetReader(csv.NewReader(b)).ReadEachWithoutHeaders(c); err != nil {
			t.Fatal(err)
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz", Frop: 1.66, Blah: &blah, SPtr: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}

	expected = Sample{Foo: "e", Bar: 3, Baz: "b", Frop: 0, Blah: &blah, SPtr: &sptr}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}
}

func Test_maybeMissingStructFields(t *testing.T) {
	structTags := []fieldInfo{
		{keys: []string{"foo"}},
		{keys: []string{"bar"}},
		{keys: []string{"baz"}},
	}
	badHeaders := []string{"hi", "mom", "bacon"}
	goodHeaders := []string{"foo", "bar", "baz"}

	// no tags to match, expect no error
	if err := maybeMissingStructFields([]fieldInfo{}, goodHeaders); err != nil {
		t.Fatal(err)
	}

	// bad headers, expect an error
	if err := maybeMissingStructFields(structTags, badHeaders); err == nil {
		t.Fatal("expected an error, but no error found")
	}

	// good headers, expect no error
	if err := maybeMissingStructFields(structTags, goodHeaders); err != nil {
		t.Fatal(err)
	}

	// extra headers, but all structtags match; expect no error
	moarHeaders := append(goodHeaders, "qux", "quux", "corge", "grault")
	if err := maybeMissingStructFields(structTags, moarHeaders); err != nil {
		t.Fatal(err)
	}

	// not all structTags match, but there's plenty o' headers; expect
	// error
	mismatchedHeaders := []string{"foo", "qux", "quux", "corgi"}
	if err := maybeMissingStructFields(structTags, mismatchedHeaders); err == nil {
		t.Fatal("expected an error, but no error found")
	}
}

func Test_maybeDoubleHeaderNames(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,foo
f,1,baz
e,3,b`)

	var samples []Sample

	// *** check maybeDoubleHeaderNames
	if err := maybeDoubleHeaderNames([]string{"foo", "BAR", "foo"}); err == nil {
		t.Fatal("maybeDoubleHeaderNames did not raise an error when a should have.")
	}

	xsvRead := NewXsvRead[Sample]()
	// *** check readTo
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	// Double header allowed, value should be of third row
	if samples[0].Foo != "baz" {
		t.Fatal("Double header allowed, value should be of third row but is not. Function called is readTo.")
	}

	b = bytes.NewBufferString(`foo,BAR,foo
f,1,baz
e,3,b`)

	xsvRead.ShouldAlignDuplicateHeadersWithStructFieldOrder = true
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	// Double header allowed, value should be of first row
	if samples[0].Foo != "f" {
		t.Fatal("Double header allowed, value should be of first row but is not. Function called is readTo.")
	}

	xsvRead.ShouldAlignDuplicateHeadersWithStructFieldOrder = false
	// Double header not allowed, should fail
	xsvRead.FailIfDoubleHeaderNames = true
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err == nil {
		t.Fatal("Double header not allowed but no error raised. Function called is readTo.")
	}

	// *** check readEach
	xsvRead.FailIfDoubleHeaderNames = false
	b = bytes.NewBufferString(`foo,BAR,foo
f,1,baz
e,3,b`)

	samples = samples[:0]
	c := make(chan Sample)
	go func() {
		if err := xsvRead.SetReader(csv.NewReader(b)).ReadEach(c); err != nil {
			t.Fatal(err)
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
	// Double header allowed, value should be of third row
	if samples[0].Foo != "baz" {
		t.Fatal("Double header allowed, value should be of third row but is not. Function called is readEach.")
	}
	// Double header not allowed, should fail
	xsvRead.FailIfDoubleHeaderNames = true
	b = bytes.NewBufferString(`foo,BAR,foo
f,1,baz
e,3,b`)

	c = make(chan Sample)
	go func() {
		if err := xsvRead.SetReader(csv.NewReader(b)).ReadEach(c); err == nil {
			t.Fatal("Double header not allowed but no error raised. Function called is readEach.")
		}
	}()
	for v := range c {
		samples = append(samples, v)
	}
}

func TestUnmarshalToCallback(t *testing.T) {
	b := bytes.NewBufferString(`first,foo,BAR,Baz,last,abc
aa,bb,11,cc,dd,ee
ff,gg,22,hh,ii,jj`)
	var samples []SkipFieldSample
	if err := NewXsvRead[SkipFieldSample]().SetByteReader(b.Bytes()).ReadToCallback(func(s SkipFieldSample) error {
		samples = append(samples, s)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}
	expected := SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "aa",
			Sample: Sample{
				Foo: "bb",
				Bar: 11,
				Baz: "cc",
			},
			Quux: "dd",
		},
		Corge: "ee",
	}
	if expected != samples[0] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
	expected = SkipFieldSample{
		EmbedSample: EmbedSample{
			Qux: "ff",
			Sample: Sample{
				Foo: "gg",
				Bar: 22,
				Baz: "hh",
			},
			Quux: "ii",
		},
		Corge: "jj",
	}
	if expected != samples[1] {
		t.Fatalf("expected first sample %v, got %v", expected, samples[1])
	}
}

// TestRenamedTypes tests for unmarshaling functions on redefined basic types.
func TestRenamedTypesUnmarshal(t *testing.T) {
	b := bytes.NewBufferString(`foo;bar
1,4;1.5
2,3;2.4`)
	var samples []RenamedSample

	// Set different csv field separator to enable comma in floats
	csvin := csv.NewReader(b)
	csvin.Comma = ';'
	if err := NewXsvRead[RenamedSample]().SetReader(csvin).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if samples[0].RenamedFloatUnmarshaler != 1.4 {
		t.Fatalf("Parsed float value wrong for renamed float64 type. Expected 1.4, got %v.", samples[0].RenamedFloatUnmarshaler)
	}
	if samples[0].RenamedFloatDefault != 1.5 {
		t.Fatalf("Parsed float value wrong for renamed float64 type without an explicit unmarshaler function. Expected 1.5, got %v.", samples[0].RenamedFloatDefault)
	}

	// Test that errors raised by UnmarshalCSV are correctly reported
	b = bytes.NewBufferString(`foo;bar
4.2;2.4`)
	csvin = csv.NewReader(b)
	csvin.Comma = ';'
	samples = samples[:0]
	if perr, _ := NewXsvRead[RenamedSample]().SetReader(csvin).ReadTo(&samples).(*csv.ParseError); perr == nil {
		t.Fatalf("Expected ParseError, got nil.")
	} else if _, ok := perr.Err.(UnmarshalError); !ok {
		t.Fatalf("Expected UnmarshalError, got %v", perr.Err)
	}
}

func (rf *RenamedFloat64Unmarshaler) UnmarshalCSV(csv string) (err error) {
	// Purely for testing purposes: Raise error on specific string
	if csv == "4.2" {
		return UnmarshalError{"Test error: Invalid float 4.2"}
	}

	// Convert , to . before parsing to create valid float strings
	converted := strings.Replace(csv, ",", ".", -1)
	var f float64
	if f, err = strconv.ParseFloat(converted, 64); err != nil {
		return err
	}
	*rf = RenamedFloat64Unmarshaler(f)
	return nil
}

// TestUnmarshalCSVWithFields test that the TestUnmarshalCSVWithFields interface to marshall all the fields works
func TestUnmarshalCSVWithFields(t *testing.T) {
	b := []byte(`foo,bar,baz,frop
bar,1,zip,3.14
baz,2,zap,4.00`)
	var samples []UnmarshalCSVWithFieldsSample
	err := NewXsvRead[UnmarshalCSVWithFieldsSample]().SetByteReader(b).ReadTo(&samples)
	if err != nil {
		t.Fatalf("UnmarshalCSVWithFields() -> UnmarshalBytes() %v", err)
	}

	tests := []struct {
		name     string
		index    int
		wantFoo  string
		wantBar  int
		wantBaz  string
		wantFrop float64
	}{
		{
			name:     "should validate index 0",
			index:    0,
			wantFoo:  "bar",
			wantBar:  1,
			wantBaz:  "zip",
			wantFrop: 314,
		},
		{
			name:     "should validate index 1",
			index:    1,
			wantFoo:  "baz",
			wantBar:  2,
			wantBaz:  "zap",
			wantFrop: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if samples[tt.index].Foo != tt.wantFoo {
				t.Fatalf("UnmarshalCSVWithFields() Index %d Foo expected %v got %v", tt.index, tt.wantFoo, samples[tt.index].Foo)
			}

			if samples[tt.index].Bar != tt.wantBar {
				t.Fatalf("UnmarshalCSVWithFields() Index %d Bar expected %v got %v", tt.index, tt.wantBar, samples[tt.index].Bar)
			}

			if samples[tt.index].Baz != tt.wantBaz {
				t.Fatalf("UnmarshalCSVWithFields() Index %d Baz expected %v got %v", tt.index, tt.wantBaz, samples[tt.index].Baz)
			}

			if samples[tt.index].Frop != tt.wantFrop {
				t.Fatalf("UnmarshalCSVWithFields() Index %d Frop expected %v got %v", tt.index, tt.wantFrop, samples[tt.index].Frop)
			}

		})
	}
}

func (u *UnmarshalCSVWithFieldsSample) UnmarshalCSVWithFields(key, value string) error {
	switch key {
	case "foo":
		u.Foo = value
	case "bar":
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		u.Bar = i
	case "baz":
		u.Baz = value
	case "frop":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		u.Frop = f * 100
	}
	return nil
}

type UnmarshalError struct {
	msg string
}

func (e UnmarshalError) Error() string {
	return e.msg
}

func TestMultipleStructTags(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,Baz
e,3,b`)

	var samples []MultiTagSample
	if err := NewXsvRead[MultiTagSample]().SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if samples[0].Foo != "b" {
		t.Fatalf("expected second tag value 'b' in multi tag struct field, got %v", samples[0].Foo)
	}

	b = bytes.NewBufferString(`foo,BAR
e,3`)

	if err := NewXsvRead[MultiTagSample]().SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if samples[0].Foo != "e" {
		t.Fatalf("wrong value in multi tag struct field, expected 'e', got %v", samples[0].Foo)
	}

	b = bytes.NewBufferString(`BAR,Baz
3,b`)

	if err := NewXsvRead[MultiTagSample]().SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}
	if samples[0].Foo != "b" {
		t.Fatal("wrong value in multi tag struct field")
	}
}

func TestStructTagSeparator(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,Baz
e,3,b`)

	var samples []TagSeparatorSample
	xsvRead := NewXsvRead[TagSeparatorSample]()
	xsvRead.TagSeparator = "|"
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}

	if samples[0].Foo != "b" {
		t.Fatal("expected second tag value in multi tag struct field.")
	}
}

func TestCustomTag(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR
e,3`)

	var samples []CustomTagSample
	xsvRead := NewXsvRead[CustomTagSample]()
	xsvRead.TagName = "custom"
	if err := xsvRead.SetReader(csv.NewReader(b)).ReadTo(&samples); err != nil {
		t.Fatal(err)
	}

	if samples[0].Foo != "e" {
		t.Fatal("wrong value in custom tag struct field")
	}
}

func TestCSVToMap(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR
4,Jose
2,Daniel
5,Vincent`)
	m, err := CSVToMap(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if m["4"] != "Jose" {
		t.Fatal("Expected Jose got", m["4"])
	}
	if m["2"] != "Daniel" {
		t.Fatal("Expected Daniel got", m["2"])
	}
	if m["5"] != "Vincent" {
		t.Fatal("Expected Vincent got", m["5"])
	}

	b = bytes.NewBufferString(`foo,BAR,Baz
e,3,b`)
	_, err = CSVToMap(bytes.NewReader(b.Bytes()))
	if err == nil {
		t.Fatal("Something went wrong")
	}
	b = bytes.NewBufferString(`foo
e`)
	_, err = CSVToMap(bytes.NewReader(b.Bytes()))
	if err == nil {
		t.Fatal("Something went wrong")
	}
}

func TestCSVToMaps(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,Baz
4,Jose,42
2,Daniel,21
5,Vincent,84`)
	m, err := NewXsvRead[interface{}]().SetReader(csv.NewReader(b)).ToMap()
	if err != nil {
		t.Fatal(err)
	}
	firstRecord := m[0]
	if firstRecord["foo"] != "4" {
		t.Fatal("Expected 4 got", firstRecord["foo"])
	}
	if firstRecord["BAR"] != "Jose" {
		t.Fatal("Expected Jose got", firstRecord["BAR"])
	}
	if firstRecord["Baz"] != "42" {
		t.Fatal("Expected 42 got", firstRecord["Baz"])
	}
	secondRecord := m[1]
	if secondRecord["foo"] != "2" {
		t.Fatal("Expected 2 got", secondRecord["foo"])
	}
	if secondRecord["BAR"] != "Daniel" {
		t.Fatal("Expected Daniel got", secondRecord["BAR"])
	}
	if secondRecord["Baz"] != "21" {
		t.Fatal("Expected 21 got", secondRecord["Baz"])
	}
	thirdRecord := m[2]
	if thirdRecord["foo"] != "5" {
		t.Fatal("Expected 5 got", thirdRecord["foo"])
	}
	if thirdRecord["BAR"] != "Vincent" {
		t.Fatal("Expected Vincent got", thirdRecord["BAR"])
	}
	if thirdRecord["Baz"] != "84" {
		t.Fatal("Expected 84 got", thirdRecord["Baz"])
	}
}

func TestCSVToMaps_OnRecord(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,Baz
4,Jose,42`)
	xsvRead := NewXsvRead[map[string]string]()
	xsvRead.OnRecord = func(record map[string]string) map[string]string {
		r2 := make(map[string]string, len(record))
		for k, v := range record {
			r2[k] = v
		}
		if r2["foo"] == "4" {
			r2["foo"] = "4-onrecord-to-maps"
		}
		return r2
	}
	m, err := xsvRead.SetReader(csv.NewReader(b)).ToMap()
	if err != nil {
		t.Fatal(err)
	}
	firstRecord := m[0]
	if firstRecord["foo"] != "4-onrecord-to-maps" {
		t.Fatal("Expected 4-onrecord-to-maps got", firstRecord["foo"])
	}
	if firstRecord["BAR"] != "Jose" {
		t.Fatal("Expected Jose got", firstRecord["BAR"])
	}
	if firstRecord["Baz"] != "42" {
		t.Fatal("Expected 42 got", firstRecord["Baz"])
	}
}

func TestCSVToMaps_FromTo(t *testing.T) {
	b := bytes.NewBufferString(`foo,BAR,Baz
4,Jose,42
2,Daniel,21
5,Vincent,84`)
	xsvRead := NewXsvRead[interface{}]()
	xsvRead.From = 2
	xsvRead.To = 3
	m, err := xsvRead.SetReader(csv.NewReader(b)).ToMap()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 2 {
		t.Fatal("Expected 2 len, but", len(m))
	}
	firstRecord := m[0]
	if firstRecord["foo"] != "2" {
		t.Fatal("Expected 2 got", firstRecord["foo"])
	}
	if firstRecord["BAR"] != "Daniel" {
		t.Fatal("Expected Daniel got", firstRecord["BAR"])
	}
	if firstRecord["Baz"] != "21" {
		t.Fatal("Expected 21 got", firstRecord["Baz"])
	}
	secondRecord := m[1]
	if secondRecord["foo"] != "5" {
		t.Fatal("Expected 5 got", secondRecord["foo"])
	}
	if secondRecord["BAR"] != "Vincent" {
		t.Fatal("Expected Vincent got", secondRecord["BAR"])
	}
	if secondRecord["Baz"] != "84" {
		t.Fatal("Expected 84 got", secondRecord["Baz"])
	}
}

func TestUnmarshalToDecoder(t *testing.T) {
	blah := 0
	sptr := "*string"
	sptr2 := ""
	b := bytes.NewBufferString(`foo,BAR,Baz,Blah,SPtr
f,1,baz,,        *string
e,3,b,,                            `)

	var samples []Sample
	if err := NewXsvRead[Sample]().SetReader(csv.NewReader(b)).Lazy().ReadToCallback(func(s Sample) error {
		samples = append(samples, s)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 sample instances, got %d", len(samples))
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz", Blah: &blah, SPtr: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}

	expected = Sample{Foo: "e", Bar: 3, Baz: "b", Blah: &blah, SPtr: &sptr2}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}
}

func TestUnmarshalWithoutHeader(t *testing.T) {
	blah := 0
	sptr := ""
	b := bytes.NewBufferString(`f,1,baz,1.66,,,
e,3,b,,,,`)

	var samples []Sample
	if err := NewXsvRead[Sample]().SetReader(csv.NewReader(b)).ReadToWithoutHeaders(&samples); err != nil {
		t.Fatal(err)
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz", Frop: 1.66, Blah: &blah, SPtr: &sptr}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}

	expected = Sample{Foo: "e", Bar: 3, Baz: "b", Frop: 0, Blah: &blah, SPtr: &sptr}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}
}

func TestUnmarshalCSVWithoutHeaders(t *testing.T) {
	// tsv input to test custom csv reader
	b := []byte("f\t1\tbaz\ne\t3\tblorp")
	r := bytes.NewReader(b)
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'

	var samples []Sample
	if err := NewXsvRead[Sample]().SetReader(csvReader).ReadToWithoutHeaders(&samples); err != nil {
		t.Fatal(err)
	}

	expected := Sample{Foo: "f", Bar: 1, Baz: "baz"}
	if !reflect.DeepEqual(expected, samples[0]) {
		t.Fatalf("expected first sample %v, got %v", expected, samples[0])
	}
	expected = Sample{Foo: "e", Bar: 3, Baz: "blorp"}
	if !reflect.DeepEqual(expected, samples[1]) {
		t.Fatalf("expected second sample %v, got %v", expected, samples[1])
	}
}

func TestDecodeDefaultValues(t *testing.T) {
	type defaultValueStruct struct {
		Foo string `csv:"foo,default=x"`
		Bar int    `csv:"bar,default=42"`
	}
	b := bytes.NewBufferString(`foo,bar
,
`)
	var out []defaultValueStruct
	if err := NewXsvRead[defaultValueStruct]().SetReader(csv.NewReader(b)).ReadTo(&out); err != nil {
		t.Fatal(err)
	}

	expected := defaultValueStruct{
		Foo: "x",
		Bar: 42,
	}
	if !reflect.DeepEqual(expected, out[0]) {
		t.Fatalf("expected second sample %v, got %v", expected, out[0])
	}
}

func TestTrimTagWhitespace(t *testing.T) {
	type whiteSpaceOptionStruct struct {
		Foo *string `csv:"foo, omitempty"`
		Bar int     `csv:"bar, default=13 "`
	}
	var out []whiteSpaceOptionStruct
	b := bytes.NewBufferString(`foo,bar
,`)
	if err := NewXsvRead[whiteSpaceOptionStruct]().SetReader(csv.NewReader(b)).ReadTo(&out); err != nil {
		t.Fatal(err)
	}
	expected := whiteSpaceOptionStruct{
		Foo: nil,
		Bar: 13,
	}

	if !reflect.DeepEqual(expected, out[0]) {
		t.Fatalf("expected sample %v, got %v", expected, out[0])
	}
}

func TestUnmarshalCSVToMap(t *testing.T) {
	b := []byte(`line	tokens
10	["PRINT", "\"Hello map!\""]
20	["GOTO", "10"]`)
	r := bytes.NewReader(b)
	csvReader := csv.NewReader(r)
	csvReader.LazyQuotes = true
	csvReader.Comma = '\t'

	var sample map[int][]string
	if err := UnmarshalCSVToMap(csvReader, &sample); err != nil {
		t.Fatal(err)
	}

	expected := map[int][]string{
		10: {"PRINT", "\"Hello map!\""},
		20: {"GOTO", "10"},
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %v, got %v", expected, sample)
	}
}

func BenchmarkCSVToMap(b *testing.B) {
	bufstring := bytes.NewBufferString(`foo,BAR
4,Jose
2,Daniel
5,Vincent`)
	for n := 0; n < b.N; n++ {
		_, err := CSVToMap(bytes.NewReader(bufstring.Bytes()))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalCSVToMap(b *testing.B) {
	bufstring := []byte(`foo,BAR
4,Jose
2,Daniel
5,Vincent`)
	for n := 0; n < b.N; n++ {
		var sample map[string]string
		r := bytes.NewReader(bufstring)
		d := csv.NewReader(r)
		err := UnmarshalCSVToMap(d, &sample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Test_readTo_nested_struct(t *testing.T) {
	b := bytes.NewBufferString(`one.boolField1,one.stringField2,two.boolField1,two.stringField2,three.boolField1,three.stringField2
false,email_one,true,email_two,false,email_three`)

	var samples []NestedSample
	err := NewXsvRead[NestedSample]().SetReader(csv.NewReader(b)).ReadTo(&samples)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expected := []NestedSample{
		{
			Inner1: InnerStruct{
				BoolIgnoreField0: false,
				BoolField1:       false,
				StringField2:     "email_one",
			},
			Inner2: InnerStruct{
				BoolIgnoreField0: false,
				BoolField1:       true,
				StringField2:     "email_two",
			},
			Inner3: NestedEmbedSample{InnerStruct{
				BoolIgnoreField0: false,
				BoolField1:       false,
				StringField2:     "email_three",
			}},
		},
	}

	if !reflect.DeepEqual(expected, samples) {
		t.Fatalf("expected \n  sample: %v\n     got: %v", expected, samples)
	}
}
