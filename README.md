# xsv
csv handling package written in go

Most of the programs related to csv generation and reading are created from code in this repository.↓

> Copyright (c) 2014 Jonathan Picques
> https://github.com/gocarina/gocsv
 
※xsv does not include gocsv.

## Getting Started

```
go get github.com/shigetaichi/xsv
```

## Usage

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