package normalizer

type KeyNormalizer interface {
	Normalize(string) string
}
