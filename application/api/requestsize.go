package api

const (
	RequestSizeTiny   RequestSize = "TINY"
	RequestSizeSmall  RequestSize = "SMALL"
	RequestSizeMedium RequestSize = "MEDIUM"
	RequestSizeLarge  RequestSize = "LARGE"
	RequestSizeXlarge RequestSize = "XLARGE"
)

type RequestSize string

func (r RequestSize) String() string {
	return string(r)
}

func (r RequestSize) IsLargerOrSame(other RequestSize) bool {
	// use reverse order of values so undefined is larger than X-large
	sizes := map[RequestSize]int{
		RequestSizeTiny:   5,
		RequestSizeSmall:  4,
		RequestSizeMedium: 3,
		RequestSizeLarge:  2,
		RequestSizeXlarge: 1,
	}

	return sizes[r] <= sizes[other]
}
