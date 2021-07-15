package normalizer

import "strings"

type RemoveStopWords struct {
	stopWords map[string]struct{}
}

func NewRemoveStopWords(stopWords []string) *RemoveStopWords {
	stopWordsSet := make(map[string]struct{})

	for _, word := range stopWords {
		stopWordsSet[word] = struct{}{}
	}

	return &RemoveStopWords{
		stopWords: stopWordsSet,
	}
}

func (n *RemoveStopWords) Normalize(key string) string {
	words := strings.Split(key, " ")
	newParts := make([]string, 0)

	for _, word := range words {
		if _, ok := n.stopWords[word]; ok {
			continue
		}

		newParts = append(newParts, word)
	}

	return strings.Join(newParts, " ")
}
