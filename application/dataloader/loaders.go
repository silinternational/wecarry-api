package dataloader

import (
	"context"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const loadersKey = "dataloaders"

type Loaders struct {
	FilesByID         FileLoader
	LocationsByID     LocationLoader
	MeetingsByID      MeetingLoader
	OrganizationsByID OrganizationLoader
	UsersByID         UserLoader
}

func getFetchFileCallback() func([]int) ([]*models.File, []error) {
	return func(ids []int) ([]*models.File, []error) {
		objects, err := models.FindFilesByIDs(ids)
		if err != nil {
			return []*models.File{}, []error{err}
		}

		objMap := map[int]models.File{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.File, len(ids))

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, nil
	}
}

func getFetchLocationCallback() func([]int) ([]*models.Location, []error) {
	return func(ids []int) ([]*models.Location, []error) {
		objects, err := models.FindLocationsByIDs(ids)
		if err != nil {
			return []*models.Location{}, []error{err}
		}

		objMap := map[int]models.Location{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.Location, len(ids))

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, nil
	}
}

func getFetchMeetingCallback() func([]int) ([]*models.Meeting, []error) {
	return func(ids []int) ([]*models.Meeting, []error) {
		objects, err := models.FindMeetingsByIDs(ids)
		if err != nil {
			return []*models.Meeting{}, []error{err}
		}

		objMap := map[int]models.Meeting{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.Meeting, len(ids))

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, nil
	}
}

func getFetchOrganizationCallback() func([]int) ([]*models.Organization, []error) {
	return func(ids []int) ([]*models.Organization, []error) {
		objects, err := models.FindOrganizationsByIDs(ids)
		if err != nil {
			return []*models.Organization{}, []error{err}
		}

		objMap := map[int]models.Organization{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.Organization, len(ids))

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, nil
	}
}

func getFetchUserCallback() func([]int) ([]*models.User, []error) {
	return func(ids []int) ([]*models.User, []error) {
		objects, err := models.FindUsersByIDs(ids)
		if err != nil {
			return []*models.User{}, []error{err}
		}

		objMap := map[int]models.User{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.User, len(ids))

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, nil
	}
}

func GetDataLoaderContext(c context.Context) context.Context {
	ctx := context.WithValue(c, loadersKey, &Loaders{
		FilesByID: FileLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchFileCallback(),
		},
		LocationsByID: LocationLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchLocationCallback(),
		},
		MeetingsByID: MeetingLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchMeetingCallback(),
		},
		OrganizationsByID: OrganizationLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchOrganizationCallback(),
		},
		UsersByID: UserLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchUserCallback(),
		},
	})

	return ctx
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
