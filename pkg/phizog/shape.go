// Package phizog is an attempt to create a more intelligent dictionary capable of identifying unmapped fields and
// assisting in analyzing data shapes and drawing inferences inference.
// The phizog store can load Shapes it has seen in the past for comparison and calculation of other statistics.
// phizog was conceived as a development aid, but could also be a production component with a little work.
package phizog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/durp/reticule/pkg/set"
	"github.com/jeremywohl/flatten"
	"github.com/sirupsen/logrus"
)

func NewStore() *Store {
	return &Store{
		shapes: make(map[string]Occurrence),
	}
}

func (s *Store) Load(shapes map[string]Occurrence) error {
	s.shapes = shapes
	logrus.Debugf("loaded %d shapes", len(s.shapes))
	return nil
}

func (s *Store) Write(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(s.shapes)
}

type Store struct {
	shapes map[string]Occurrence
}

type Occurrence struct {
	Shape Shape
	Count int
}

func (s *Store) Add(shape Shape) {
	v, ok := s.shapes[shape.ID()]
	if !ok {
		s.shapes[shape.ID()] = Occurrence{
			Shape: shape,
			Count: 1,
		}
		return
	}
	v.Count++
	s.shapes[shape.ID()] = v
}

func (s *Store) Dump() string {
	ids := make(set.Strings, len(s.shapes))
	for k := range s.shapes {
		ids.Add(k)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids.Slice() {
		var o Occurrence
		out = append(out, fmt.Sprintf("%d shape %s %v", o.Count, id, o.Shape))
	}
	return strings.Join(out, "\n")
}

func (s *Store) AddShape(name string, raw interface{}) error {
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			err := s.AddShape(name, item)
			if err != nil {
				return err
			}
		}
	case map[string]interface{}:
		flat, err := flatten.Flatten(v, "", flatten.DotStyle)
		if err != nil {
			return err
		}
		keys := make(set.Strings, len(flat))
		for k := range flat {
			keys.Add(k)
		}
		s.Add(Shape{
			Name: name,
			Keys: keys,
		})
	default:
		logrus.Errorf("unhandled type %T", v)
	}
	return nil
}

func (s Shape) ID() string {
	hasher := sha256.New()
	hasher.Write([]byte(strings.Join(s.Keys.Slice(), ",")))
	return hex.EncodeToString(hasher.Sum(nil))
}

type Shape struct {
	Name string
	Keys set.Strings
}
