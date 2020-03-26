package models

type PostSize string

const (
	PostSizeTiny   PostSize = "TINY"
	PostSizeSmall  PostSize = "SMALL"
	PostSizeMedium PostSize = "MEDIUM"
	PostSizeLarge  PostSize = "LARGE"
	PostSizeXlarge PostSize = "XLARGE"
)

func (p PostSize) String() string {
	return string(p)
}

func (p PostSize) isLargerOrSame(other PostSize) bool {
	sizes := map[PostSize]int{
		PostSizeTiny:   1,
		PostSizeSmall:  2,
		PostSizeMedium: 3,
		PostSizeLarge:  4,
		PostSizeXlarge: 5,
	}
	return sizes[p] >= sizes[other]
}
