package database

import (
	"github.com/mendelgusmao/scoredb/lib/fuzzymap"
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	cmap "github.com/orcaman/concurrent-map/v2"

	"fmt"
)

type Database struct {
	collections cmap.ConcurrentMap[*fuzzymap.FuzzyMap[any]]
}

func NewDatabase() *Database {
	return &Database{
		collections: cmap.New[*fuzzymap.FuzzyMap[any]](),
	}
}

func (s *Database) CollectionExists(collectionName string) bool {
	return s.collections.Has(collectionName)
}

func (s *Database) CreateCollection(collectionName string, config Configuration, documents []Document) error {
	if s.CollectionExists(collectionName) {
		return fmt.Errorf(collectionAlreadyExistsError, collectionName)
	}

	fuzzyMap := fuzzymap.New[any](
		fuzzymap.FuzzyMapConfig{
			UseLevenshtein: config.UseLevenshtein,
			GramSizeLower:  config.GramSizeLower,
			GramSizeUpper:  config.GramSizeUpper,
			MinScore:       config.MinScore,
			SetConfiguration: normalizer.SetConfiguration{
				Synonyms:      config.Synonyms,
				StopWords:     config.StopWords,
				Transliterate: config.Transliterate,
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
