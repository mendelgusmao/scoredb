package normalizer

type NormalizerSet struct {
	normalizers []KeyNormalizer
}

func NewNormalizerSet(normalizers ...KeyNormalizer) *NormalizerSet {
	return &NormalizerSet{
		normalizers: normalizers,
	}
}

func (n *NormalizerSet) Normalize(key string) string {
	for _, normalizer := range n.normalizers {
		key = normalizer.Normalize(key)
	}

	return key
}
