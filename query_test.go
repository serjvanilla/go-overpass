package overpass

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		input string
		want  Result
	}{
		{
			`{"elements":[{"type":"way","id":1,
				"bounds":{"minlat":-37.9,"minlon":144.6,"maxlat":-37.8,"maxlon":144.7}
			}]}`,
			Result{
				Count: 1,
				Ways: map[int64]*Way{1: {
					Bounds: Box{Min: Point{-37.9, 144.6}, Max: Point{-37.8, 144.7}},
				}},
			},
		},
		{
			`{"elements":[{"type":"way","id":1,
				"geometry":[{"lat":-37.9,"lon":144.6},{"lat":-37.8,"lon":144.7}]
			}]}`,
			Result{
				Count: 1,
				Ways: map[int64]*Way{1: {
					Geometry: []Point{{-37.9, 144.6}, {-37.8, 144.7}},
				}},
			},
		},
		{
			`{"elements":[{"type":"relation","id":1,
				"bounds":{"minlat":-37.9,"minlon":144.6,"maxlat":-37.8,"maxlon":144.7}
			}]}`,
			Result{
				Count: 1,
				Relations: map[int64]*Relation{1: {
					Bounds: Box{Min: Point{-37.9, 144.6}, Max: Point{-37.8, 144.7}},
				}},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			got, err := unmarshal([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}

			if tc.want.Nodes == nil {
				tc.want.Nodes = map[int64]*Node{}
			} else {
				for id, n := range tc.want.Nodes {
					n.Meta.ID = id
				}
			}
			if tc.want.Ways == nil {
				tc.want.Ways = map[int64]*Way{}
			} else {
				for id, w := range tc.want.Ways {
					w.Meta.ID = id
					if w.Nodes == nil {
						w.Nodes = []*Node{}
					}
					if w.Geometry == nil {
						w.Geometry = []Point{}
					}
				}
			}
			if tc.want.Relations == nil {
				tc.want.Relations = map[int64]*Relation{}
			} else {
				for id, r := range tc.want.Relations {
					r.Meta.ID = id
					if r.Members == nil {
						r.Members = []RelationMember{}
					}
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("%v != %v", got, tc.want)
			}
		})
	}
}
