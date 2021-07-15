package normalizer

import "strings"

type ReplaceSynonyms struct {
	synonyms map[string]string
}

func NewReplaceSynonyms(synonyms map[string]string) *ReplaceSynonyms {
	return &ReplaceSynonyms{
		synonyms: synonyms,
	}
}

func (n *ReplaceSynonyms) Normalize(key string) string {
	words := strings.Split(key, " ")

	for index, word := range words {
		if replacement, ok := n.synonyms[word]; ok {
			words[index] = replacement
		}
	}

	return strings.Join(words, " ")
}
