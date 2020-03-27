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
	// use reverse order of values so undefined is larger than X-large
	sizes := map[PostSize]int{
		PostSizeTiny:   5,
		PostSizeSmall:  4,
		PostSizeMedium: 3,
		PostSizeLarge:  2,
		PostSizeXlarge: 1,
	}

	return sizes[p] <= sizes[other]
}
