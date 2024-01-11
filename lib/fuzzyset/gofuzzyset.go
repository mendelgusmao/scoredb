package gofuzzyset

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// This is a re-implementation of the javascript fuzzyset thingy
type FuzzySet struct {
	// key is gram size, value is "item"
	itemsByGramSize map[int][]item

	// For each gram track index and gramcount
	// Key is gram, value is something
	matchDict map[string][]uint16

	// key = normalized value and value = value (e.g. exactSet["hello"] = "Hello"
	exactSet map[string]string

	useLevenshtein bool
	gramSizeLower  int
	gramSizeUpper  int
	minScore       float64
}

// For each word in the dataset, "normaize" it and calculate the vector (root of sum of squares of gram counts)
type item struct {
	normalizedValue string
	vectorNormal    float64
}

type Match struct {
	Word  string
	Score float64
}

func New(data []string, useLevenshtein bool, gramSizeLower int, gramSizeUpper int, minScore float64) *FuzzySet {
	f := FuzzySet{
		useLevenshtein: useLevenshtein,
		gramSizeLower:  gramSizeLower,
		gramSizeUpper:  gramSizeUpper,
		minScore:       minScore,
	}

	// Initialize items structs
	f.itemsByGramSize = make(map[int][]item)
	f.matchDict = make(map[string][]uint16)
	f.exactSet = make(map[string]string)

	for gramSize := gramSizeLower; gramSize <= gramSizeUpper; gramSize++ {
		f.itemsByGramSize[gramSize] = make([]item, 0)
	}

	// Add data to FuzzySet set
	for i := range data {
		f.Add(data[i])
	}

	return &f
}

func (f FuzzySet) Add(value string) {
	normalizedValue := normalizeStr(value)

	// If this normaized value is in the exact set already, then ignore
	if _, found := f.exactSet[normalizedValue]; found {
		return
	}

	for gramSize := f.gramSizeLower; gramSize <= f.gramSizeUpper; gramSize++ {
		items := f.itemsByGramSize[gramSize]
		index := len(items)

		gramsByCount := gramCounter(value, gramSize)
		sumOfSquareGramCounts := 0.0

		for gram, gramCount := range gramsByCount {
			sumOfSquareGramCounts = sumOfSquareGramCounts + float64(gramCount*gramCount)
			reference := uint16(index)<<8 | uint16(gramCount)

			if _, found := f.matchDict[gram]; found {
				f.matchDict[gram] = append(f.matchDict[gram], uint16(reference))
			} else {
				f.matchDict[gram] = []uint16{reference}
			}
		}

		vectorNormal := math.Sqrt(sumOfSquareGramCounts)
		items = append(items, item{normalizedValue: normalizedValue, vectorNormal: vectorNormal})
		f.itemsByGramSize[gramSize] = items
		f.exactSet[normalizedValue] = value
	}
}

/*
Search for a value with a score of at least minScore...return the found value along w/ the score
*/
func (f FuzzySet) Get(value string) []Match {
	results := make([]Match, 0)

	// Check for exact match
	if exactMatch, found := f.exactSet[normalizeStr(value)]; found {
		return []Match{{Word: exactMatch, Score: 1.0}}
	}

	// start with high gram size and if there are no results, go to lower gram sizes
	for gramSize := f.gramSizeUpper; gramSize >= f.gramSizeLower; gramSize-- {
		results = f.findMatchesForGramSize(value, gramSize)

		if len(results) > 0 {
			break
		}
	}

	return results
}

func (f FuzzySet) findMatchesForGramSize(value string, gramSize int) []Match {
	var results []Match
	matches := make(map[uint16]uint16, 0)

	normalizedValue := normalizeStr(value)

	gramCountsByGram := gramCounter(normalizedValue, gramSize)
	sumOfSquareGramCounts := 0.0

	for gram, gramCount := range gramCountsByGram {
		sumOfSquareGramCounts = sumOfSquareGramCounts + float64(gramCount*gramCount)

		if gramMatchDict, found := f.matchDict[gram]; found {
			for i := 0; i < len(gramMatchDict); i++ {
				index := gramMatchDict[i] >> 8
				otherGramCount := gramMatchDict[i] & 0xFF

				if _, found := matches[index]; found {
					matches[index] = matches[index] + gramCount*otherGramCount
				} else {
					matches[index] = gramCount * otherGramCount
				}
			}
		}
	}

	if len(matches) == 0 {
		return results
	}

	vectorNormal := math.Sqrt(sumOfSquareGramCounts)

	for i := range matches {
		score := matches[i]
		item := f.itemsByGramSize[gramSize][i]
		normScore := float64(score) / (vectorNormal * item.vectorNormal)

		results = append(results, Match{Word: item.normalizedValue, Score: normScore})
	}

	sort.Sort(byScore(results))

	// If desired, "levenshtein-ize" the scores and re-sort
	if f.useLevenshtein {
		newResults := make([]Match, 0)

		for i := range results {
			newResults = append(newResults, Match{Score: distance(results[i].Word, normalizedValue), Word: results[i].Word})
		}

		results = newResults
		sort.Sort(byScore(results))
	}

	// Filter results by min score
	newResults := make([]Match, 0)

	for i := range results {
		if results[i].Score >= f.minScore {
			newResults = append(newResults, results[i])
		}
	}

	return newResults
}

func normalizeStr(str string) string {
	return strings.ToLower(str)
}

func levenshtein(str1, str2 string) int {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}

	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}

			column[y] = min(column[y]+1, column[y-1]+1, lastkey+incr)
			lastkey = oldkey
		}
	}

	return column[s1len]
}

// return an edit distance from 0 to 1
func distance(str1, str2 string) float64 {
	if str1 == "" {
		return 0
	}

	if str2 == "" {
		return 0
	}

	d := levenshtein(str1, str2)

	if len(str1) > len(str2) {
		return 1.0 - float64(d)/float64(len(str1))
	} else {
		return 1.0 - float64(d)/float64(len(str2))
	}
}

var nonWordRE = regexp.MustCompile("/[^a-zA-Z0-9\u00C0-\u00FF, ]+/g")

func iterateGrams(value string, gramSize int) []string {
	grams := make([]string, 0)

	simplified := fmt.Sprintf("-%v-", nonWordRE.ReplaceAllString(strings.ToLower(value), ""))
	lenDiff := gramSize - len(simplified)

	if lenDiff > 0 {
		simplified = simplified + strings.Repeat("-", lenDiff)
	}

	for i := 0; i < len(simplified)-gramSize+1; i++ {
		gram := simplified[i : i+gramSize]
		grams = append(grams, gram)
	}

	return grams
}

// Results = map with grams as key and number of occurances as values
func gramCounter(value string, gramSize int) map[string]uint16 {
	results := make(map[string]uint16)
	grams := iterateGrams(value, gramSize)

	for i := range grams {
		if _, found := results[grams[i]]; found {
			results[grams[i]] = results[grams[i]] + 1
		} else {
			results[grams[i]] = 1
		}
	}

	return results
}

func min(things ...int) int {
	currentMin := math.MaxInt64

	for _, thing := range things {
		if thing < currentMin {
			currentMin = thing
		}
	}

	return currentMin
}

type byScore []Match

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].Score > a[j].Score }
