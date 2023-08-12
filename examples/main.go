package main

import (
	"fmt"
	"os"
	"sync"
	"time"
	"xsv"
)

type NotUsed struct {
	Name string
}

type Client struct { // Our example struct, you can use "-" to ignore a field
	ID            string   `csv:"client_id"`
	Name          string   `csv:"client_name"`
	Age           string   `csv:"client_age"`
	NotUsedString string   `csv:"-"`
	NotUsedStruct NotUsed  `csv:"-"`
	Address1      Address  `csv:"addr1"`
	Address2      Address  //`csv:"addr2"` will use Address2 in header
	Employed      DateTime `csv:"employed"`
}
type Address struct {
	Street string `csv:"street"`
	City   string `csv:"city"`
}

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.String(), nil
}

// You could also use the standard Stringer interface
func (date DateTime) String() string {
	return date.Time.Format("20060201")
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("20060201", csv)
	return err
}

func main() {
	// Create clients
	clients := []*Client{
		{ID: "12", Name: "John", Age: "21",
			Address1: Address{"Street 1", "City1"},
			Address2: Address{"Street 2", "City2"},
			Employed: DateTime{time.Date(2022, 11, 04, 12, 0, 0, 0, time.UTC)},
		},
		{ID: "13", Name: "Fred",
			Address1: Address{`Main "Street" 1`, "City1"}, // show quotes in value
			Address2: Address{"Main Street 2", "City2"},
			Employed: DateTime{time.Date(2022, 11, 04, 13, 0, 0, 0, time.UTC)},
		},
		{ID: "14", Name: "James", Age: "32",
			Address1: Address{"Center Street 1", "City1"},
			Address2: Address{"Center Street 2", "City2"},
			Employed: DateTime{time.Date(2022, 11, 04, 14, 0, 0, 0, time.UTC)},
		},
		{ID: "15", Name: "Danny",
			Address1: Address{"State Street 1", "City1"},
			Address2: Address{"State Street 2", "City2"},
			Employed: DateTime{time.Date(2022, 11, 04, 15, 0, 0, 0, time.UTC)},
		},
	}
	// Create an empty clients file
	clientsFile, err := os.OpenFile("clients.csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	xsvWrite := xsv.NewXSVWrite[*Client]()
	err = xsvWrite.SetFileWriter(clientsFile).Write(clients)
	if err != nil {
		return
	}

	// WriteFromChan
	var wg sync.WaitGroup
	clientsChan := make(chan *Client)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i + 1
		go func() {
			defer wg.Done()
			v := Client{
				ID: fmt.Sprintf("%v", i), Name: "Danny",
				Address1: Address{"State Street 1", "City1"},
				Address2: Address{"State Street 2", "City2"},
				Employed: DateTime{time.Date(2022, 11, 04, 15, 0, 0, 0, time.UTC)},
			}
			clientsChan <- &v
		}()
	}

	go func() {
		wg.Wait()
		close(clientsChan)
	}()

	// Create an empty clientsFromChan file
	clientsFromChanFile, err := os.OpenFile("clients-from-chan.csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	f := xsvWrite.SetFileWriter(clientsFromChanFile)
	f.OmitHeaders = true
	err = f.WriteFromChan(clientsChan)
	if err != nil {
		panic(err)
	}
}
