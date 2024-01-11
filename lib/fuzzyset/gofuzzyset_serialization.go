package gofuzzyset

import (
	"bytes"
	"fmt"

	msgpack "github.com/vmihailenco/msgpack/v5"
)

type FuzzySetRepresentation struct {
	ItemsByGramSize map[int][]item
	MatchDict       map[string][]uint16
	ExactSet        map[string]string
	UseLevenshtein  bool
	GramSizeLower   int
	GramSizeUpper   int
	MinScore        float64
}

type ItemRepresentation struct {
	NormalizedValue string
	VectorNormal    float64
}

func (f *FuzzySet) MarshalMsgpack() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	enc := msgpack.NewEncoder(buffer)
	fuzzySetRepr := FuzzySetRepresentation{
		ItemsByGramSize: f.itemsByGramSize,
		MatchDict:       f.matchDict,
		ExactSet:        f.exactSet,
		UseLevenshtein:  f.useLevenshtein,
		GramSizeLower:   f.gramSizeLower,
		GramSizeUpper:   f.gramSizeUpper,
		MinScore:        f.minScore,
	}

	if err := enc.Encode(fuzzySetRepr); err != nil {
		return nil, fmt.Errorf("[FuzzySet] %v", err)
	}

	return buffer.Bytes(), nil
}

func (f *FuzzySet) UnmarshalMsgpack(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := msgpack.NewDecoder(buffer)

	fuzzySetRepr := FuzzySetRepresentation{
		ItemsByGramSize: make(map[int][]item),
		MatchDict:       make(map[string][]uint16),
		ExactSet:        make(map[string]string),
	}

	if err := dec.Decode(&fuzzySetRepr); err != nil {
		return fmt.Errorf("[FuzzySet] %v", err)
	}

	f.itemsByGramSize = fuzzySetRepr.ItemsByGramSize
	f.matchDict = fuzzySetRepr.MatchDict
	f.exactSet = fuzzySetRepr.ExactSet
	f.useLevenshtein = fuzzySetRepr.UseLevenshtein
	f.gramSizeLower = fuzzySetRepr.GramSizeLower
	f.gramSizeUpper = fuzzySetRepr.GramSizeUpper
	f.minScore = fuzzySetRepr.MinScore

	return nil
}

func (i *item) MarshalMsgpack() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	enc := msgpack.NewEncoder(buffer)
	itemRepr := []interface{}{
		i.normalizedValue,
		i.vectorNormal,
	}

	if err := enc.Encode(itemRepr); err != nil {
		return nil, fmt.Errorf("[FuzzySet.Item] %v", err)
	}

	return buffer.Bytes(), nil
}

func (i *item) UnmarshalMsgpack(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := msgpack.NewDecoder(buffer)
	fuzzySetRepr := []interface{}{}

	if err := dec.Decode(&fuzzySetRepr); err != nil {
		return fmt.Errorf("[FuzzySet.Item] %v", err)
	}

	i.normalizedValue = fuzzySetRepr[0].(string)
	i.vectorNormal = fuzzySetRepr[1].(float64)

	return nil
}
