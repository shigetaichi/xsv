package xsv

import (
	"bytes"
	"encoding/csv"
	"testing"
)

func TestUnmarshalListOfStructsAfterMarshal(t *testing.T) {

	type Additional struct {
		Value string
	}

	type Option struct {
		Additional []*Additional
		Key        string
	}

	inData := []*Option{
		{
			Key: "test",
		},
	}

	// First, marshal our test data to a CSV format
	buffer := new(bytes.Buffer)
	innerWriter := csv.NewWriter(buffer)
	innerWriter.Comma = '|'
	xsvWrite := NewXsvWrite[*Option]()
	err := xsvWrite.SetWriter(innerWriter).Write(inData)
	if err != nil {
		t.Fatalf("Error marshalling data to CSV: %#v", err)
	}

	if string(buffer.Bytes()) != "Additional|Key\nnull|test\n" {
		t.Fatalf("Marshalled data had an unexpected form of %s", buffer.Bytes())
	}

	// Next, attempt to unmarshal our test data from a CSV format
	var outData []*Option
	innerReader := csv.NewReader(buffer)
	innerReader.Comma = '|'
	if err := NewXsvRead[*Option]().SetReader(innerReader).ReadTo(&outData); err != nil {
		t.Fatalf("Error unmarshalling data from CSV: %#v", err)
	}

	// Finally, verify the data
	if len(outData) != 1 {
		t.Fatalf("Data expected to have one entry, had %d entries", len(outData))
	} else if len(outData[0].Additional) != 0 {
		t.Fatalf("Data Additional field expected to be empty, had length of %d", len(outData[0].Additional))
	} else if outData[0].Key != "test" {
		t.Fatalf("Data Key field did not contain expected value, had %q", outData[0].Key)
	}
}
