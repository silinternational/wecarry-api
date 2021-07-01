package conversions

import (
	"context"
	"errors"

	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

// converts list of model.Request into api.RequestAbridged
func ConvertRequestsAbridged(ctx context.Context, requests []models.Request) ([]api.RequestAbridged, error) {
	output := make([]api.RequestAbridged, len(requests))

	for i, request := range requests {
		var err error
		output[i], err = ConvertRequestToAPITypeAbridged(ctx, request)
		if err != nil {
			return []api.RequestAbridged{}, err
		}
	}

	return output, nil
}

// converts model.Request into api.Request
func ConvertRequestToAPIType(ctx context.Context, request models.Request) (api.Request, error) {
	var output api.Request
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.Request{}, err
	}
	output.ID = request.UUID

	// Hydrate nested request fields for api.RequestAbridged
	tx := models.Tx(ctx)
	createdBy, err := hydrateRequestCreatedBy(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.CreatedBy = &createdBy

	origin, err := hydrateRequestOrigin(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.Origin = &origin

	destination, err := hydrateRequestDestination(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.Destination = &destination

	provider, err := hydrateProvider(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.Provider = &provider

	photo, err := hydrateRequestPhoto(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.Photo = &photo

	// TODO: hydrate other nested request fields after reconciling api.Request struct with the UI field list
	organization, err := hydrateRequestOrganization(ctx, request, tx)
	if err != nil {
		return api.Request{}, err
	}
	output.Organization = &organization

	return output, nil
}

// converts model.Request into api.RequestAbridged
func ConvertRequestToAPITypeAbridged(ctx context.Context, request models.Request) (api.RequestAbridged, error) {
	var output api.RequestAbridged
	if err := api.ConvertToOtherType(request, &output); err != nil {
		err = errors.New("error converting request to api.request: " + err.Error())
		return api.RequestAbridged{}, err
	}
	output.ID = request.UUID

	// Hydrate nested request fields
	tx := models.Tx(ctx)
	createdBy, err := hydrateRequestCreatedBy(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.CreatedBy = &createdBy

	origin, err := hydrateRequestOrigin(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Origin = &origin

	destination, err := hydrateRequestDestination(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Destination = &destination

	provider, err := hydrateProvider(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Provider = &provider

	photo, err := hydrateRequestPhoto(ctx, request, tx)
	if err != nil {
		return api.RequestAbridged{}, err
	}
	output.Photo = &photo

	return output, nil
}

func hydrateRequestCreatedBy(ctx context.Context, request models.Request, tx *pop.Connection) (api.User, error) {
	createdBy, _ := request.GetCreator(tx)
	outputCreatedBy, err := ConvertUser(ctx, *createdBy)
	if err != nil {
		err = errors.New("error converting request created_by user: " + err.Error())
		return api.User{}, err
	}
	return outputCreatedBy, nil
}

func hydrateRequestOrigin(ctx context.Context, request models.Request, tx *pop.Connection) (api.Location, error) {
	origin, err := request.GetOrigin(tx)
	if err != nil {
		err = errors.New("error converting request origin: " + err.Error())
		return api.Location{}, err
	}
	var outputOrigin api.Location
	if err := api.ConvertToOtherType(origin, &outputOrigin); err != nil {
		err = errors.New("error converting origin to api.Location: " + err.Error())
		return api.Location{}, err
	}
	return outputOrigin, nil
}

func hydrateRequestDestination(ctx context.Context, request models.Request, tx *pop.Connection) (api.Location, error) {
	destination, err := request.GetDestination(tx)
	if err != nil {
		err = errors.New("error converting request destination: " + err.Error())
		return api.Location{}, err
	}
	var outputDestination api.Location
	if err := api.ConvertToOtherType(destination, &outputDestination); err != nil {
		err = errors.New("error converting destination to api.Location: " + err.Error())
		return api.Location{}, err
	}
	return outputDestination, nil
}

func hydrateProvider(ctx context.Context, request models.Request, tx *pop.Connection) (api.User, error) {
	provider, err := request.GetProvider(tx)
	if err != nil {
		err = errors.New("error converting request provider: " + err.Error())
		return api.User{}, err
	}
	var outputProvider api.User
	if err := api.ConvertToOtherType(provider, &outputProvider); err != nil {
		err = errors.New("error converting provider to api.User: " + err.Error())
		return api.User{}, err
	}
	if provider != nil {
		outputProvider.ID = provider.UUID
	}
	return outputProvider, nil
}

func hydrateRequestPhoto(ctx context.Context, request models.Request, tx *pop.Connection) (api.File, error) {
	photo, err := request.GetPhoto(tx)
	if err != nil {
		err = errors.New("error converting request photo: " + err.Error())
		return api.File{}, err
	}
	var outputPhoto api.File
	if err := api.ConvertToOtherType(photo, &outputPhoto); err != nil {
		err = errors.New("error converting photo to api.File: " + err.Error())
		return api.File{}, err
	}
	if photo != nil {
		outputPhoto.ID = photo.UUID
	}
	return outputPhoto, nil
}

func hydrateRequestOrganization(ctx context.Context, request models.Request, tx *pop.Connection) (api.Organization, error) {
	// Hydrate the request's organization
	organization, err := request.GetOrganization(tx)
	if err != nil {
		err = errors.New("error converting request organization: " + err.Error())
		return api.Organization{}, err
	}
	var outputOrg api.Organization
	if err := api.ConvertToOtherType(organization, &outputOrg); err != nil {
		err = errors.New("error converting organization to api.Organization: " + err.Error())
		return api.Organization{}, err
	}
	outputOrg.ID = organization.UUID
	return outputOrg, nil
}
