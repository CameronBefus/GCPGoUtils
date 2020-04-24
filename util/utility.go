package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
)

var viperInitialized = false

// InitializeViper -
func InitializeViper(fn string) bool {
	if !viperInitialized {
		viper.SetConfigName(fn)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		// viper.AddConfigPath("..\\explorer")
		// viper.AddConfigPath("..\\..\\explorer")
		viper.AutomaticEnv()
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic("No configuration file in use")
		}
		viperInitialized = true
	}
	return viperInitialized
}

func GetConfigName() string {
	return viper.ConfigFileUsed()
}

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

type CustomLogFormat struct{}

func (lf *CustomLogFormat) Format(entry *log.Entry) ([]byte, error) {
	if entry.Level.String() == "info" {
		return []byte(fmt.Sprintf("%s \n", entry.Message)), nil
	}
	return []byte(fmt.Sprintf("%s: %s \n", entry.Level, entry.Message)), nil
}

func JSONMarshalNoEscape(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
