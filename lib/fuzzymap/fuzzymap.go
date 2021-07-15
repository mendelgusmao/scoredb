package fuzzymap

import (
	"github.com/mendelgusmao/scoredb/lib/fuzzymap/normalizer"
	"github.com/mendelgusmao/scoredb/lib/set"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/samprakos/gofuzzyset"
)

type FuzzyMap struct {
	candidates    cmap.ConcurrentMap // map[string, []interface{}]
	fuzzySet      *gofuzzyset.FuzzySet
	keyNormalizer normalizer.KeyNormalizer
}

type Match struct {
	Score   int8        `json:"score"`
	Content interface{} `json:"content"`
}

func New(useLevenshtein bool, gramSizeLower int, gramSizeUpper int, minScore float64, keyNormalizer normalizer.KeyNormalizer) *FuzzyMap {
	return &FuzzyMap{
		candidates: cmap.New(), // map[string, []interface{}]
		fuzzySet: gofuzzyset.New(
			[]string{},
			useLevenshtein,
			gramSizeLower,
			gramSizeUpper,
			minScore,
		),
		keyNormalizer: keyNormalizer,
	}
}

func (fm *FuzzyMap) Add(key string, item interface{}) {
	key = fm.normalizeKey(key)
	fm.fuzzySet.Add(key)
	fm.add(key, item)
}

func (fm *FuzzyMap) AddExact(key string, item interface{}) {
	key = fm.normalizeKey(key)
	fm.add(key, item)
}

func (fm *FuzzyMap) add(key string, item interface{}) {
	candidates, ok := fm.candidates.Get(key)

	if !ok {
		candidates = set.New(item)
	} else {
		candidates.(*set.Set).Insert(item)
	}

	fm.candidates.Set(key, candidates)
}

func (fm *FuzzyMap) Get(key string) []Match {
	if len(key) == 0 {
		return []Match{}
	}

	key = fm.normalizeKey(key)

	if exactMatches, ok := fm.candidates.Get(key); ok {
		return fm.expandCandidates(exactMatches.(*set.Set), 100)
	}

	return fm.fuzzyFind(key)
}

func (fm *FuzzyMap) add(key string, item interface{}) {
	candidates, ok := fm.candidates.Get(key)

	if !ok {
		candidates = set.New(item)
	} else {
		candidates.(*set.Set).Insert(item)
	}

	fm.candidates.Set(key, candidates)
}

func (fm *FuzzyMap) normalizeKey(key string) string {
	if fm.keyNormalizer != nil {
		return fm.keyNormalizer.Normalize(key)
	}

	return key
}

func (fm *FuzzyMap) expandCandidates(candidates *set.Set, score int8) []Match {
	matches := make([]Match, candidates.Len())
	index := 0

	candidates.Do(func(candidate interface{}) {
		matches[index] = Match{score, candidate}
		index++
	})

	return matches
}

func (fm *FuzzyMap) fuzzyFind(key string) []Match {
	scoredKeys := fm.fuzzySet.Get(key)
	matches := make([]Match, 0)

	for _, scoredKey := range scoredKeys {
		score := int8(scoredKey.Score * 100)
		candidates, ok := fm.candidates.Get(scoredKey.Word)

		if !ok {
			continue
		}

		matches = append(matches, fm.expandCandidates(candidates.(*set.Set), score)...)
	}

	return matches
}
