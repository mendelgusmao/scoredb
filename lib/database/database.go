package database

import (
	"bytes"
	"encoding/json"

	"github.com/mendelgusmao/scoredb/lib/fuzzymap"
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	cmap "github.com/orcaman/concurrent-map/v2"

	"fmt"
)

type Database struct {
	collections cmap.ConcurrentMap[*fuzzymap.FuzzyMap[any]]
}

type DatabaseRepresentation struct {
	Collections map[string]*fuzzymap.FuzzyMap[any]
}

func NewDatabase() *Database {
	return &Database{
		collections: cmap.New[*fuzzymap.FuzzyMap[any]](),
	}
}

func (s *Database) CollectionExists(collectionName string) bool {
	return s.collections.Has(collectionName)
}

func (s *Database) CreateCollection(collectionName string, fuzzySetConfig FuzzySetConfiguration, documents []Document) error {
	if s.CollectionExists(collectionName) {
		return fmt.Errorf(collectionAlreadyExistsError, collectionName)
	}

	fuzzyMap := fuzzymap.New[any](
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

func (s *Database) UpdateCollection(collectionName string, documents []Document) error {
	fuzzyMap, ok := s.collections.Get(collectionName)

	if !ok {
		return fmt.Errorf(collectionDoesntExistError, "UpdateCollection", collectionName)
	}

	s.addDocumentsToFuzzyMap(fuzzyMap, documents)

	return nil
}

func (s *Database) Query(collectionName, key string) ([]fuzzymap.Match[any], error) {
	fuzzyMap, ok := s.collections.Get(collectionName)

	if !ok {
		return []fuzzymap.Match[any]{}, fmt.Errorf(collectionDoesntExistError, "Get", collectionName)
	}

	return fuzzyMap.Get(key), nil
}

func (s *Database) RemoveCollection(collectionName string) error {
	if !s.collections.Has(collectionName) {
		return fmt.Errorf(collectionDoesntExistError, "RemoveCollection", collectionName)
	}

	s.collections.Remove(collectionName)

	return nil
}

func (s *Database) addDocumentsToFuzzyMap(fuzzyMap *fuzzymap.FuzzyMap[any], documents []Document) {
	for _, document := range documents {
		for _, key := range document.Keys {
			fuzzyMap.Add(key, document.Content)
		}

		for _, key := range document.ExactKeys {
			fuzzyMap.AddExact(key, document.Content)
		}
	}
}

func (s *Database) MarshalJSON() ([]byte, error) {
	collections := make(map[string]*fuzzymap.FuzzyMap[any])

	for collectionTuple := range s.collections.IterBuffered() {
		collections[collectionTuple.Key] = collectionTuple.Val
	}

	databaseRepr := &DatabaseRepresentation{
		Collections: collections,
	}

	buffer := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buffer)

	if err := enc.Encode(databaseRepr); err != nil {
		return nil, fmt.Errorf("[Set] %v", err)
	}

	return buffer.Bytes(), nil
}

func (s *Database) UnmarshalJSON(input []byte) error {
	buffer := bytes.NewBuffer(input)
	dec := json.NewDecoder(buffer)

	databaseRepr := DatabaseRepresentation{
		Collections: make(map[string]*fuzzymap.FuzzyMap[any]),
	}

	if err := dec.Decode(&databaseRepr); err != nil {
		return fmt.Errorf("[Set] %v", err)
	}

	for key, value := range databaseRepr.Collections {
		s.collections.Set(key, value)
	}

	return nil
}
