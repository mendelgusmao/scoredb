package fuzzyset

import (
	"bytes"
	"fmt"

	"github.com/mendelgusmao/gofuzzyset"
	msgpack "github.com/vmihailenco/msgpack/v5"
)

type FuzzySet struct {
	gofuzzyset.FuzzySet `gob:"-"`
}

func (f *FuzzySet) MarshalMsgpack() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	enc := msgpack.NewEncoder(buffer)
	fuzzySetRepr := f.Export()

	if err := enc.Encode(fuzzySetRepr); err != nil {
		return nil, fmt.Errorf("[FuzzySet] %v", err)
	}

	return buffer.Bytes(), nil
}

func (f *FuzzySet) UnmarshalMsgpack(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := msgpack.NewDecoder(buffer)

	fuzzySetRepr := gofuzzyset.NewFuzzySetRepresentation()

	if err := dec.Decode(&fuzzySetRepr); err != nil {
		return fmt.Errorf("[FuzzySet] %v", err)
	}

	f.Import(fuzzySetRepr)

	return nil
}
