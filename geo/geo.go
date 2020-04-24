package geo

import (
	"fmt"
	geoj "github.com/paulmach/go.geojson"
)

type Bounds struct {
	minLat  float64
	maxLat  float64
	minLong float64
	maxLong float64
}

func ExtractBounds(geometry string) (Bounds, error) {
	var b Bounds
	bg1, e1 := geoj.UnmarshalGeometry([]byte(geometry))
	if e1 != nil {
		return b, e1
	}
	if bg1.IsPolygon() {
		doPolygon(bg1.Polygon[0], &b)
	} else if bg1.IsMultiPolygon() {
		for _, m := range bg1.MultiPolygon {
			doPolygon(m[0], &b)
		}
	} else {
		return b, fmt.Errorf(`unexpected geometry type : %s`, bg1.Type)
	}
	return b, nil
}

func doPolygon(geo [][]float64, b *Bounds) {
	if b.minLat == 0.0 {
		b.minLat = geo[0][1]
		b.minLong = geo[0][0]
		b.maxLat = b.minLat
		b.maxLong = b.minLong
	}
	for _, pt := range geo {
		if pt[0] < b.maxLong {
			b.maxLong = pt[0]
		}
		if pt[0] > b.minLong {
			b.minLong = pt[0]
		}
		if pt[1] < b.minLat {
			b.minLat = pt[1]
		}
		if pt[1] > b.maxLat {
			b.maxLat = pt[1]
		}
	}
}

//func updateBounds() {
//	q := `REPLACE INTO county_bounds (state_id, county_id, min_lat, max_lat, min_long, max_long) VALUES (?,?,?,?,?,?)`
//	for _, c := range counties {
//		if c.min_lat > 0 {
//			_, err := db.MySQL.Exec(q, c.stateID, c.countyID, c.min_lat, c.max_lat, c.min_long, c.max_long)
//			if err != nil {
//				log.Error(err)
//			}
//		}
//	}
//}
