package util

import (
	"reflect"
	"testing"
)

func Test_uniqueInts(t *testing.T) {
	tests := []struct {
		name string
		args []int
		want []int
	}{
		{`a`, []int{1}, []int{1}},
		{`b`, []int{1, 2, 2, 3}, []int{1, 2, 3}},
		{`c`, []int{}, []int{}},
		{`d`, []int{1, 2, 1, 2}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueInts(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqueInts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntSliceToCSV(t *testing.T) {
	tests := []struct {
		name string
		args []int
		want string
	}{
		{`a`, []int{1, 2, 3}, `1,2,3`},
		{`b`, []int{1}, `1`},
		{`c`, []int{}, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntSliceToCSV(tt.args); got != tt.want {
				t.Errorf("IntSliceToCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshalNoEscape(t *testing.T) {

	type ts struct {
		Content string
	}
	tests := []struct {
		name    string
		args    ts
		want    string
		wantErr bool
	}{
		{`a`, ts{`hello`}, `{"Content":"hello"}` + "\n", false},
		{`b`, ts{`Sanford & Son`}, `{"Content":"Sanford & Son"}` + "\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONMarshalNoEscape(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONMarshalNoEscape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gs := string(got)
			if gs != tt.want {
				t.Errorf(`JSONMarshalNoEscape() got = "%s", want "%s"`, gs, tt.want)
			}
		})
	}
}
