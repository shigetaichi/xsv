package xsv

import (
	"testing"
)

func TestXsvRead_checkFromTo(t *testing.T) {
	type Arg struct {
		From int
		To   int
	}
	type testCase[T any] struct {
		name    string
		arg     Arg
		wantErr bool
	}
	tests := []testCase[Arg]{
		{
			name:    "default",
			arg:     Arg{From: 1, To: -1},
			wantErr: false,
		},
		{
			name:    "only 1 line",
			arg:     Arg{From: 1, To: 1},
			wantErr: false,
		},
		{
			name:    "great and small reversals",
			arg:     Arg{From: 3, To: 2},
			wantErr: true,
		},
		{
			name:    "unsigned int on From",
			arg:     Arg{From: -1, To: 2},
			wantErr: true,
		},
		{
			name:    "unsigned int on To",
			arg:     Arg{From: 1, To: -2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xsvRead := NewXsvRead[any]()
			xsvRead.From = tt.arg.From
			xsvRead.To = tt.arg.To
			if err := xsvRead.checkFromTo(); (err != nil) != tt.wantErr {
				t.Errorf("checkFromTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
