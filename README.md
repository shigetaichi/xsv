# xsv
csv handling package written in go

Most of the programs related to csv generation and reading are created from code in this repository.↓

> Copyright (c) 2014 Jonathan Picques
> https://github.com/gocarina/gocsv

※　xsv does not include gocsv.

## 🚧Under Construction
there is a lot of things to do before release.

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
	
	xsvWrite := xsv.NewXSVWrite[*Client](clients)
	err = xsvWrite.WriteToFile(clientsFile)
	if err != nil {
		return
	}
}

```