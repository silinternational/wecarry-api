package models

import "github.com/silinternational/wecarry-api/api"

type RequestSize string

const (
	RequestSizeTiny   RequestSize = "TINY"
	RequestSizeSmall  RequestSize = "SMALL"
	RequestSizeMedium RequestSize = "MEDIUM"
	RequestSizeLarge  RequestSize = "LARGE"
	RequestSizeXlarge RequestSize = "XLARGE"
)

func (r RequestSize) String() string {
	return string(r)
}

func (r RequestSize) isLargerOrSame(other RequestSize) bool {
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

func GetRequestSizeFromAPISize(apiSize api.RequestSize) RequestSize {
	sizeMap := map[api.RequestSize]RequestSize{
		api.RequestSizeTiny:   RequestSizeTiny,
		api.RequestSizeSmall:  RequestSizeSmall,
		api.RequestSizeMedium: RequestSizeMedium,
		api.RequestSizeLarge:  RequestSizeLarge,
		api.RequestSizeXlarge: RequestSizeXlarge,
	}
	return sizeMap[apiSize]
}
