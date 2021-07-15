package database

import (
	"github.com/mendelgusmao/scoredb/lib/fuzzymap"
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	cmap "github.com/orcaman/concurrent-map"

	"fmt"
)

type Database struct {
	collections cmap.ConcurrentMap // map[string, FuzzyMap]
}

func NewDatabase() *Database {
	return &Database{
		collections: cmap.New(),
	}
}

func (s *Database) CreateCollection(collectionName string, config Configuration, documents []Document) error {
	if _, ok := s.collections.Get(collectionName); ok {
		return fmt.Errorf(collectionAlreadyExistsError, collectionName)
	}

	fuzzyMap := fuzzymap.New(
		config.UseLevenshtein,
		config.GramSizeLower,
		config.GramSizeUpper,
		config.MinScore,
		buildNormalizerSet(&config),
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

	s.addDocumentsToFuzzyMap(fuzzyMap.(*fuzzymap.FuzzyMap), documents)

	return nil
}

func (s *Database) Query(collectionName, key string) ([]fuzzymap.Match, error) {
	fuzzyMap, ok := s.collections.Get(collectionName)

	if !ok {
		return []fuzzymap.Match{}, fmt.Errorf(collectionDoesntExistError, "Get", collectionName)
	}

	return fuzzyMap.(*fuzzymap.FuzzyMap).Get(key), nil
}

func (s *Database) RemoveCollection(collectionName string) error {
	if !s.collections.Has(collectionName) {
		return fmt.Errorf(collectionDoesntExistError, "RemoveCollection", collectionName)
	}

	s.collections.Remove(collectionName)

	return nil
}

func (s *Database) addDocumentsToFuzzyMap(fuzzyMap *fuzzymap.FuzzyMap, documents []Document) {
	for _, document := range documents {
		for _, key := range document.Keys {
			fuzzyMap.Add(key, document.Content)
		}

		for _, key := range document.ExactKeys {
			fuzzyMap.AddExact(key, document.Content)
		}
	}
}

func buildNormalizerSet(config *Configuration) *normalizer.NormalizerSet {
	normalizers := []normalizer.KeyNormalizer{
		normalizer.NewBasicFilter(),
	}

	if config.Transliterate {
		normalizers = append(normalizers, normalizer.NewTransliterate())
	}

	if len(config.StopWords) > 0 {
		normalizers = append(normalizers, normalizer.NewRemoveStopWords(config.StopWords))
	}

	if len(config.Synonyms) > 0 {
		normalizers = append(normalizers, normalizer.NewReplaceSynonyms(config.Synonyms))
	}

	return normalizer.NewNormalizerSet(normalizers...)
}
