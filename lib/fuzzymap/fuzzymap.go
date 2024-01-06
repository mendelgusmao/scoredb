package fuzzymap

import (
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	"github.com/mendelgusmao/scoredb/lib/set"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/samprakos/gofuzzyset"
)

type FuzzyMap[V any] struct {
	candidates       cmap.ConcurrentMap[*set.Set[V]]
	fuzzySet         *gofuzzyset.FuzzySet
	normalizerConfig normalizer.SetConfiguration
	keyNormalizer    normalizer.KeyNormalizer
}

type FuzzyMapConfig struct {
	UseLevenshtein bool
	GramSizeLower  int
	GramSizeUpper  int
	MinScore       float64
	normalizer.SetConfiguration
}

type Match[V any] struct {
	Score   int8 `json:"score"`
	Content V    `json:"content"`
}

func New[V any](config FuzzyMapConfig) *FuzzyMap[V] {
	fuzzyMap := &FuzzyMap[V]{
		candidates: cmap.New[*set.Set[V]](),
		fuzzySet: gofuzzyset.New(
			[]string{},
			config.UseLevenshtein,
			config.GramSizeLower,
			config.GramSizeUpper,
			config.MinScore,
		),
		normalizerConfig: config.SetConfiguration,
	}

	fuzzyMap.ApplyNormalizer()

	return fuzzyMap
}

func (fm *FuzzyMap[V]) ApplyNormalizer() {
	fm.keyNormalizer = normalizer.NewDefaultSet(fm.normalizerConfig)
}

func (fm *FuzzyMap[V]) Add(key string, item V) {
	key = fm.normalizeKey(key)
	fm.fuzzySet.Add(key)
	fm.add(key, item)
}

func (fm *FuzzyMap[V]) AddExact(key string, item V) {
	key = fm.normalizeKey(key)
	fm.add(key, item)
}

func (fm *FuzzyMap[V]) Get(key string) []Match[V] {
	if len(key) == 0 {
		return []Match[V]{}
	}

	key = fm.normalizeKey(key)

	if exactMatches, ok := fm.candidates.Get(key); ok {
		return fm.expandCandidates(exactMatches, 100)
	}

	return fm.fuzzyFind(key)
}

func (fm *FuzzyMap[V]) add(key string, item V) {
	candidates, ok := fm.candidates.Get(key)

	if !ok {
		candidates = set.New(item)
	} else {
		candidates.Insert(item)
	}

	fm.candidates.Set(key, candidates)
}

func (fm *FuzzyMap[V]) normalizeKey(key string) string {
	if fm.keyNormalizer != nil {
		return fm.keyNormalizer.Normalize(key)
	}

	return key
}

func (fm *FuzzyMap[V]) expandCandidates(candidates *set.Set[V], score int8) []Match[V] {
	matches := make([]Match[V], candidates.Len())
	index := 0

	candidates.Do(func(candidate V) {
		matches[index] = Match[V]{score, candidate}
		index++
	})

	return matches
}

func (fm *FuzzyMap[V]) fuzzyFind(key string) []Match[V] {
	scoredKeys := fm.fuzzySet.Get(key)
	matches := make([]Match[V], 0)

	for _, scoredKey := range scoredKeys {
		score := int8(scoredKey.Score * 100)
		candidates, ok := fm.candidates.Get(scoredKey.Word)

		if !ok {
			continue
		}

		matches = append(matches, fm.expandCandidates(candidates, score)...)
	}

	return matches
}
