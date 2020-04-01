package dataloader

import (
	"context"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const loadersKey = "dataloaders"

type Loaders struct {
	MeetingByID      MeetingLoader
	OrganizationByID OrganizationLoader
	UserByID         UserLoader
}

func getFetchMeetingCallback() func([]int) ([]*models.Meeting, []error) {
	return func(ids []int) ([]*models.Meeting, []error) {
		gotErrors := []error{}

		objectsByID := map[int]*models.Meeting{}
		for _, id := range ids {
			obj := models.Meeting{}
			if _, ok := objectsByID[id]; ok {
				continue
			}
			err := obj.FindByID(id)
			if err != nil {
				gotErrors = append(gotErrors, err)
				continue
			}
			objectsByID[obj.ID] = &obj
		}

		objects := make([]*models.Meeting, len(ids))
		for i, id := range ids {
			objects[i] = objectsByID[id]
		}

		return objects, nil
	}
}

func getFetchOrganizationCallback() func([]int) ([]*models.Organization, []error) {
	return func(ids []int) ([]*models.Organization, []error) {
		gotErrors := []error{}

		objectsByID := map[int]*models.Organization{}
		for _, id := range ids {
			obj := models.Organization{}
			if _, ok := objectsByID[id]; ok {
				continue
			}
			err := obj.FindByID(id)
			if err != nil {
				gotErrors = append(gotErrors, err)
				continue
			}
			objectsByID[obj.ID] = &obj
		}

		objects := make([]*models.Organization, len(ids))
		for i, id := range ids {
			objects[i] = objectsByID[id]
		}

		return objects, nil
	}
}

func getFetchUserCallback() func([]int) ([]*models.User, []error) {
	return func(ids []int) ([]*models.User, []error) {
		gotErrors := []error{}

		objectsByID := map[int]*models.User{}
		for _, id := range ids {
			obj := models.User{}
			if _, ok := objectsByID[id]; ok {
				continue
			}
			err := obj.FindByID(id)
			if err != nil {
				gotErrors = append(gotErrors, err)
				continue
			}
			objectsByID[obj.ID] = &obj
		}

		objects := make([]*models.User, len(ids))
		for i, id := range ids {
			objects[i] = objectsByID[id]
		}

		return objects, nil
	}
}

func GetDataLoaderContext(c context.Context) context.Context {
	ctx := context.WithValue(c, loadersKey, &Loaders{
		MeetingByID: MeetingLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchMeetingCallback(),
		},
		OrganizationByID: OrganizationLoader{
			maxBatch: domain.DataLoaderMaxBatch,
			wait:     domain.DataLoaderWaitMilliSeconds,
			fetch:    getFetchOrganizationCallback(),
		},
		UserByID: UserLoader{
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
