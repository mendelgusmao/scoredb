package database

import (
	"bytes"

	"github.com/mendelgusmao/scoredb/lib/fuzzymap"
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	cmap "github.com/orcaman/concurrent-map/v2"
	msgpack "github.com/vmihailenco/msgpack/v5"

	"fmt"
)

type Database[T any] struct {
	collections cmap.ConcurrentMap[*fuzzymap.FuzzyMap[T]]
}

type DatabaseRepresentation[T any] struct {
	Collections map[string]*fuzzymap.FuzzyMap[T]
}

func NewDatabase[T any]() *Database[T] {
	return &Database[T]{
		collections: cmap.New[*fuzzymap.FuzzyMap[T]](),
	}
}

func (s *Database[T]) CollectionExists(collectionName string) bool {
	return s.collections.Has(collectionName)
}

func (s *Database[T]) CreateCollection(collectionName string, fuzzySetConfig FuzzySetConfiguration, documents []Document[T]) error {
	if s.CollectionExists(collectionName) {
		return fmt.Errorf(collectionAlreadyExistsError, collectionName)
	}

	fuzzyMap := fuzzymap.New[T](
		fuzzymap.FuzzyMapConfig{
			UseLevenshtein: fuzzySetConfig.UseLevenshtein,
			GramSizeLower:  fuzzySetConfig.GramSizeLower,
			GramSizeUpper:  fuzzySetConfig.GramSizeUpper,
			MinScore:       fuzzySetConfig.MinScore,
			SetConfiguration: normalizer.SetConfiguration{
				Synonyms:      fuzzySetConfig.Synonyms,
				StopWords:     fuzzySetConfig.StopWords,
				Transliterate: fuzzySetConfig.Transliterate,
			},
		},
	)

	s.addDocumentsToFuzzyMap(fuzzyMap, documents)

	s.collections.Set(collectionName, fuzzyMap)

	return nil
}

func (s *Database[T]) UpdateCollection(collectionName string, documents []Document[T]) error {
	fuzzyMap, ok := s.collections.Get(collectionName)

	if !ok {
		return fmt.Errorf(collectionDoesntExistError, "UpdateCollection", collectionName)
	}

	s.addDocumentsToFuzzyMap(fuzzyMap, documents)

	return nil
}

func (s *Database[T]) Query(collectionName, key string) ([]fuzzymap.Match[T], error) {
	fuzzyMap, ok := s.collections.Get(collectionName)

	if !ok {
		return []fuzzymap.Match[T]{}, fmt.Errorf(collectionDoesntExistError, "Get", collectionName)
	}

	return fuzzyMap.Get(key), nil
}

func (s *Database[T]) RemoveCollection(collectionName string) error {
	if !s.collections.Has(collectionName) {
		return fmt.Errorf(collectionDoesntExistError, "RemoveCollection", collectionName)
	}

	s.collections.Remove(collectionName)

	return nil
}

func (s *Database[T]) addDocumentsToFuzzyMap(fuzzyMap *fuzzymap.FuzzyMap[T], documents []Document[T]) {
	for _, document := range documents {
		for _, key := range document.Keys {
			fuzzyMap.Add(key, document.Content)
		}

		for _, key := range document.ExactKeys {
			fuzzyMap.AddExact(key, document.Content)
		}
	}
}

func (s *Database[T]) MarshalMsgpack() ([]byte, error) {
	collections := make(map[string]*fuzzymap.FuzzyMap[T])

	for collectionTuple := range s.collections.IterBuffered() {
		collections[collectionTuple.Key] = collectionTuple.Val
	}

	databaseRepr := &DatabaseRepresentation[T]{
		Collections: collections,
	}

	buffer := bytes.NewBuffer(nil)
	enc := msgpack.NewEncoder(buffer)

	if err := enc.Encode(databaseRepr); err != nil {
		return nil, fmt.Errorf("[Set] %v", err)
	}

	return buffer.Bytes(), nil
}

func (s *Database[T]) UnmarshalMsgpack(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := msgpack.NewDecoder(buffer)

	databaseRepr := DatabaseRepresentation[T]{
		Collections: make(map[string]*fuzzymap.FuzzyMap[T]),
	}

	if err := dec.Decode(&databaseRepr); err != nil {
		return fmt.Errorf("[Set] %v", err)
	}

	for key, value := range databaseRepr.Collections {
		s.collections.Set(key, value)
	}

	return nil
}
