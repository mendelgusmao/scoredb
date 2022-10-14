package normalizer

type Set struct {
	normalizers []KeyNormalizer
}

type SetConfiguration struct {
	Synonyms      map[string]string `json:"synonyms,omitempty"`
	StopWords     []string          `json:"stopWords,omitempty"`
	Transliterate bool              `json:"transliterate,omitempty"`
}

func NewDefaultSet(config *SetConfiguration) *Set {
	normalizers := []KeyNormalizer{
		NewBasicFilter(),
	}

	if config.Transliterate {
		normalizers = append(normalizers, NewTransliterate())
	}

	if len(config.StopWords) > 0 {
		normalizers = append(normalizers, NewRemoveStopWords(config.StopWords))
	}

	if len(config.Synonyms) > 0 {
		normalizers = append(normalizers, NewReplaceSynonyms(config.Synonyms))
	}

	return NewSet(normalizers...)
}

func NewSet(normalizers ...KeyNormalizer) *Set {
	return &Set{
		normalizers: normalizers,
	}
}

func (n *Set) Normalize(key string) string {
	for _, normalizer := range n.normalizers {
		key = normalizer.Normalize(key)
	}

	return key
}
