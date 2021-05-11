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

func convertErrToSlice(err error) []error {
	if err == nil {
		return nil
	}
	return []error{err}
}

func getFetchFileCallback() func([]int) ([]*models.File, []error) {
	return func(ids []int) ([]*models.File, []error) {
		objects := models.Files{}
		err := objects.FindByIDs(models.DB, ids)
		if len(objects) == 0 {
			return []*models.File{}, convertErrToSlice(err)
		}

		objMap := map[int]models.File{}
		for _, o := range objects {
			objMap[o.ID] = o
		}

		objPtrs := make([]*models.File, len(ids))
		errors := make([]error, len(ids))
		if err != nil {
			errors[0] = err
		}

		for i, id := range ids {
			if obj, ok := objMap[id]; ok {
				objPtrs[i] = &obj
			}
		}

		return objPtrs, convertErrToSlice(err)
	}
}

func getFetchLocationCallback() func([]int) ([]*models.Location, []error) {
	return func(ids []int) ([]*models.Location, []error) {
		objects := models.Locations{}
		err := objects.FindByIDs(ids)
		if len(objects) == 0 {
			return []*models.Location{}, convertErrToSlice(err)
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

		return objPtrs, convertErrToSlice(err)
	}
}

func getFetchMeetingCallback() func([]int) ([]*models.Meeting, []error) {
	return func(ids []int) ([]*models.Meeting, []error) {
		objects := models.Meetings{}
		err := objects.FindByIDs(ids)
		if len(objects) == 0 {
			return []*models.Meeting{}, convertErrToSlice(err)
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

		return objPtrs, convertErrToSlice(err)
	}
}

func getFetchOrganizationCallback() func([]int) ([]*models.Organization, []error) {
	return func(ids []int) ([]*models.Organization, []error) {
		objects := models.Organizations{}
		err := objects.FindByIDs(ids)
		if len(objects) == 0 {
			return []*models.Organization{}, convertErrToSlice(err)
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

		return objPtrs, convertErrToSlice(err)
	}
}

func getFetchUserCallback() func([]int) ([]*models.User, []error) {
	return func(ids []int) ([]*models.User, []error) {
		objects := models.Users{}
		err := objects.FindByIDs(ids)
		if len(objects) == 0 {
			return []*models.User{}, convertErrToSlice(err)
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

		return objPtrs, convertErrToSlice(err)
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
