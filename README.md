# xsv

csv handling package written in go

[![Licence](https://img.shields.io/github/license/shigetaichi/xsv)](https://github.com/shigetaichi/xsv/blob/main/LICENSE)
[![Code Size](https://img.shields.io/github/languages/code-size/shigetaichi/xsv)](https://github.com/shigetaichi/xsv)
[![Release](https://img.shields.io/github/v/release/shigetaichi/xsv)](https://github.com/shigetaichi/xsv/releases)
[![Github Stars](https://img.shields.io/github/stars/shigetaichi/xsv)](https://github.com/shigetaichi/xsv/stargazers)

Most of the programs related to csv generation and reading are created from code in this repository.↓

> Copyright (c) 2014 Jonathan Picques
> https://github.com/gocarina/gocsv

※xsv does not include gocsv.

## 🚀Getting Started

```
go get github.com/shigetaichi/xsv
```

## 🔨Usage

```go
package main

import (
	"os"
	"xsv"
)

type Client struct { // Our example struct, you can use "-" to ignore a field
	ID            string `csv:"client_id"`
	Name          string `csv:"client_name"`
	Age           string `csv:"client_age"`
	NotUsedString string `csv:"-"`
}

func main() {
	clients := []*Client{
		{ID: "12", Name: "John", Age: "21"},
		{ID: "13", Name: "Fred"},
		{ID: "14", Name: "James", Age: "32"},
		{ID: "15", Name: "Danny"},
	}

	// Create an empty clients file
	clientsFile, err := os.OpenFile("clients.csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	// instancing xsvWrite struct
	xsvWrite := xsv.NewXsvWrite[*Client]()
	// change some preferences
	xsvWrite.OmitHeaders = true
	// set writer and write!
	err = xsvWrite.SetFileWriter(clientsFile).Write(clients)
	if err != nil {
		return
	}

	// instancing xsvRead struct
	xsvRead := xsv.NewXsvRead[*Client]()
	// change some preferences
	xsvRead.TagName = "xsv"
	// set reader and read!
	var clientOutput []*Client
	err = xsvRead.SetFileReader(clientsFile).ReadTo(&clientOutput)
	if err != nil {
		return
	}
}

```

## 🛠️Details

### XsvWrite
- **TagName**: `string`
    - Key in the struct field's tag to scan
- **TagSeparator**: `string`
    - Separator string for multiple CSV tags in struct fields
- **OmitHeaders**: `bool`
    - Whether to output headers to CSV or not
- **SelectedColumns**: `[]string`
    - Slice of field names (which is set in "TagName" tag) to output
- **SortOrder**: `[]uint`
    - Column sort order
- **HeaderModifier**: `map[string]string`
    - Map to dynamically change headers
- **OnRecord** `func(T) T`
    - Callback function to be called on each record

### XsvRead
- **TagName**: `string`
    - Key in the struct field's tag to scan
- **TagSeparator**: `string`
    - Separator string for multiple CSV tags in struct fields
- **FailIfUnmatchedStructTags**: `bool`
    - Indicates whether it is considered an error when there is an unmatched struct tag.
- **FailIfDoubleHeaderNames**: `bool`
    - Indicates whether it is considered an error when a header name is repeated in the CSV header.
- **ShouldAlignDuplicateHeadersWithStructFieldOrder**: `bool`
    - Indicates whether duplicate CSV headers should be aligned per their order in the struct definition.
- **OnRecord** `func(T) T`
    - Callback function to be called on each record
- **NameNormalizer**: `Normalizer(func(string) string)`
    - Normalizer is a function that takes and returns a string. It is applied to struct and header field values before they are compared. It can be used to alter names for comparison. For instance, you could allow case-insensitive matching or convert '-' to '_'.
- **ErrorHandler**: `ErrorHandler(func(*csv.ParseError) bool)`
    - ErrorHandler is a function that takes an error and handles it. It can be used to log errors or to panic.
