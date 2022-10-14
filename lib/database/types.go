package database

type Configuration struct {
	UseLevenshtein bool              `json:"useLevenshtein,omitempty"`
	GramSizeLower  int               `json:"gramSizeLower,omitempty"`
	GramSizeUpper  int               `json:"gramSizeUpper,omitempty"`
	MinScore       float64           `json:"minScore,omitempty"`
	Synonyms       map[string]string `json:"synonyms,omitempty"`
	StopWords      []string          `json:"stopWords,omitempty"`
	Transliterate  bool              `json:"transliterate,omitempty"`
}

type Document struct {
	Keys      []string `json:"keys"`
	ExactKeys []string `json:"exactKeys"`
	Content   any      `json:"content"`
}

const (
	collectionDoesntExistError   = "database.Database.%s: collection `%s` doesn't exist"
	collectionAlreadyExistsError = "database.Database.CreateCollection: collection `%s` already exists"
)
