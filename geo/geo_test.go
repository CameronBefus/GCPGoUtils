package geo

import (
	"reflect"
	"testing"
)

func TestExtractBounds(t *testing.T) {
	var null Bounds
	tests := []struct {
		name    string
		arg     string
		want    Bounds
		wantErr bool
	}{
		{`a`, ``, null, true},
		{`b`, `{"type":"Polygon","coordinates":[[[-1.5,5.5],[-1.4,5.6],[-1.3,5.7],[-1.2,5.5],[-1.35,5.4]]]}`, Bounds{5.4, 5.7, -1.2, -1.5}, false},
		{`c`, `{"type":"MultiPolygon","coordinates":[[[[-1.6,5.6],[-1.5,5.7],[-1.4,5.8],[-1.3,5.6],[-1.45,5.5]]],[[[-1.5,5.5],[-1.4,5.6],[-1.3,5.7],[-1.2,5.5],[-1.35,5.4]]]]}`, Bounds{5.4, 5.8, -1.2, -1.6}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractBounds(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractBounds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractBounds() got = %v, want %v", got, tt.want)
			}
		})
	}
}
