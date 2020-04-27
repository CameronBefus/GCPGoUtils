package util

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
)

// UniqueInts returns a unique subset of the int slice provided.
func UniqueInts(input []int) []int {
	u := make([]int, 0, len(input))
	m := make(map[int]bool)
	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// IntSliceToCSV - convert a slice of integers into a CSV string
func IntSliceToCSV(ns []int) string {
	if len(ns) == 0 {
		return ""
	}
	// Appr. 3 chars per num plus the comma.
	estimate := len(ns) * 3
	b := make([]byte, 0, estimate)
	for x, n := range ns {
		if x > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendInt(b, int64(n), 10)
	}
	return string(b)
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func JSONMarshalNoEscape(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// FileExists - takes a filename, returns bool
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
