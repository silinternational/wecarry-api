package gqlgen

import (
	"context"
	"errors"
	"fmt"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/dataloader"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Request returns the request resolver. It is required by GraphQL
func (r *Resolver) Request() RequestResolver {
	return &requestResolver{r}
}

func (r *queryResolver) Request(ctx context.Context, id *string) (*models.Request, error) {
	if id == nil {
		return nil, errors.New("ID must not be null")
	}

	cUser := models.CurrentUser(ctx)
	domain.NewExtra(ctx, "requestUUID", *id)

	request := models.Request{}
	if err := request.FindByUUID(models.Tx(ctx), *id); err != nil {
		return &request, domain.ReportError(ctx, err, "ViewRequest.Error")
	}

	if request.ID != 0 && cUser.CanViewRequest(ctx, request) {
		return &request, nil
	}

	return &models.Request{}, domain.ReportError(ctx, errors.New("user not allowed to view request"),
		"ViewRequest.NotFound")
}

type requestResolver struct{ *Resolver }

// ID resolves the `ID` property of the request query. It provides the UUID instead of the autoincrement ID.
func (r *requestResolver) ID(ctx context.Context, obj *models.Request) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// CreatedBy resolves the `createdBy` property of the request query. It retrieves the related record from the database.
func (r *requestResolver) CreatedBy(ctx context.Context, obj *models.Request) (*PublicProfile, error) {
	if obj == nil {
		return &PublicProfile{}, nil
	}

	creator, err := dataloader.For(ctx).UsersByID.Load(obj.CreatedByID)
	if err != nil {
		return &PublicProfile{}, domain.ReportError(ctx, err, "GetRequestCreator")
	}

	return getPublicProfile(ctx, creator), nil
}

// Provider resolves the `provider` property of the request query. It retrieves the related record from the database.
func (r *requestResolver) Provider(ctx context.Context, obj *models.Request) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.ProviderID.Valid {
		return nil, nil
	}

	provider, err := dataloader.For(ctx).UsersByID.Load(obj.ProviderID.Int)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestProvider")
	}

	return getPublicProfile(ctx, provider), nil
}

// PotentialProviders resolves the `potentialProviders` property of the request query,
// retrieving the related records from the database.
func (r *requestResolver) PotentialProviders(ctx context.Context, obj *models.Request) ([]PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	providers, err := obj.GetPotentialProviders(ctx, models.CurrentUser(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetPotentialProviders")
	}

	profiles := getPublicProfiles(ctx, providers)
	return profiles, nil
}

// Organization resolves the `organization` property of the request query. It retrieves the related record from the
// database.
func (r *requestResolver) Organization(ctx context.Context, obj *models.Request) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organization, err := dataloader.For(ctx).OrganizationsByID.Load(obj.OrganizationID)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestOrganization")
	}
	return organization, nil
}

// Description resolves the `description` property, converting a nulls.String to a *string.
func (r *requestResolver) Description(ctx context.Context, obj *models.Request) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.Description), nil
}

// NeededBefore resolves the `neededBefore` property of the request query, converting a nulls.Time to a *string.
func (r *requestResolver) NeededBefore(ctx context.Context, obj *models.Request) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	return models.GetStringFromNullsTime(obj.NeededBefore), nil
}

// CompletedOn resolves the `completedOn` property of the request query, converting a nulls.Time to a *string.
func (r *requestResolver) CompletedOn(ctx context.Context, obj *models.Request) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	return models.GetStringFromNullsTime(obj.CompletedOn), nil
}

// Destination resolves the `destination` property of the request query, retrieving the related record from the database.
func (r *requestResolver) Destination(ctx context.Context, obj *models.Request) (*models.Location, error) {
	if obj == nil {
		return &models.Location{}, nil
	}

	destination, err := dataloader.For(ctx).LocationsByID.Load(obj.DestinationID)
	if err != nil {
		return &models.Location{}, domain.ReportError(ctx, err, "GetRequestDestination")
	}

	return destination, nil
}

// Origin resolves the `origin` property of the request query, retrieving the related record from the database.
func (r *requestResolver) Origin(ctx context.Context, obj *models.Request) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.OriginID.Valid {
		return nil, nil
	}

	origin, err := dataloader.For(ctx).LocationsByID.Load(obj.OriginID.Int)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestOrigin")
	}

	return origin, nil
}

// Actions resolves the `actions` property of the request query, retrieving the related records from the database.
func (r *requestResolver) Actions(ctx context.Context, obj *models.Request) ([]string, error) {
	if obj == nil {
		return []string{}, nil
	}

	user := models.CurrentUser(ctx)

	actions, err := obj.GetCurrentActions(ctx, user)
	if err != nil {
		return actions, domain.ReportError(ctx, err, "GetRequestActions")
	}

	return actions, nil
}

// Threads resolves the `threads` property of the request query, retrieving the related records from the database.
func (r *requestResolver) Threads(ctx context.Context, obj *models.Request) ([]models.Thread, error) {
	if obj == nil {
		return nil, nil
	}

	user := models.CurrentUser(ctx)
	threads, err := obj.GetThreads(ctx, user)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestThreads")
	}

	return threads, nil
}

// URL resolves the `url` property of the request query, converting nulls.String to a *string
func (r *requestResolver) URL(ctx context.Context, obj *models.Request) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.URL), nil
}

// Kilograms resolves the `kilograms` property of the request query as a pointer to a float64
func (r *requestResolver) Kilograms(ctx context.Context, obj *models.Request) (*float64, error) {
	if obj == nil {
		return nil, nil
	}
	if !obj.Kilograms.Valid {
		return nil, nil
	}

	return &obj.Kilograms.Float64, nil
}

// Photo retrieves the file attached as the primary photo
func (r *requestResolver) Photo(ctx context.Context, obj *models.Request) (*models.File, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.FileID.Valid {
		return nil, nil
	}

	photo, err := dataloader.For(ctx).FilesByID.Load(obj.FileID.Int)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestPhoto")
	}

	return photo, nil
}

// PhotoID retrieves UUID of the file attached as the Request photo
func (r *requestResolver) PhotoID(ctx context.Context, obj *models.Request) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.FileID.Valid {
		return nil, nil
	}

	photoID, err := obj.GetPhotoID(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetUserPhotoID")
	}

	return photoID, nil
}

// Files retrieves the list of files attached to the request, not including the primary photo
func (r *requestResolver) Files(ctx context.Context, obj *models.Request) ([]models.File, error) {
	if obj == nil {
		return nil, nil
	}
	files, err := obj.GetFiles(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestFiles")
	}

	return files, nil
}

// Meeting resolves the `meeting` property of the request query, retrieving the related record from the database.
func (r *requestResolver) Meeting(ctx context.Context, obj *models.Request) (*models.Meeting, error) {
	if obj == nil {
		return nil, nil
	}

	if !obj.MeetingID.Valid {
		return nil, nil
	}

	meeting, err := dataloader.For(ctx).MeetingsByID.Load(obj.MeetingID.Int)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequestMeeting")
	}
	return meeting, nil
}

// IsEditable indicates whether the user is allowed to edit the request
func (r *requestResolver) IsEditable(ctx context.Context, obj *models.Request) (bool, error) {
	if obj == nil {
		return false, nil
	}
	cUser := models.CurrentUser(ctx)
	return obj.IsEditable(ctx, cUser)
}

// Requests resolves the `requests` query
func (r *queryResolver) Requests(ctx context.Context, destination, origin *LocationInput, searchText *string) (
	[]models.Request, error) {

	requests := models.Requests{}
	cUser := models.CurrentUser(ctx)

	filter := models.RequestFilterParams{
		Destination: convertOptionalLocation(destination),
		Origin:      convertOptionalLocation(origin),
		SearchText:  searchText,
	}
	err := requests.FindByUser(ctx, cUser, filter)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetRequests")
	}

	return requests, nil
}

//// Request resolves the `request` query
//func (r *queryResolver) Request(ctx context.Context, id *string) (*models.Request, error) {
//	if id == nil {
//		return nil, nil
//	}
//	var request models.Request
//	cUser := models.CurrentUser(ctx)
//	if err := request.FindByUserAndUUID(ctx, cUser, *id); err != nil {
//		extras := map[string]interface{}{
//			"user": cUser.UUID,
//		}
//		return nil, domain.ReportError(ctx, err, "GetRequest", extras)
//	}
//
//	return &request, nil
//}

// convertGqlRequestInputToDBRequest takes a `RequestInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Request` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertGqlRequestInputToDBRequest(ctx context.Context, input requestInput, currentUser models.User) (models.Request, error) {
	request := models.Request{}

	if input.ID != nil {
		if err := request.FindByUUID(models.Tx(ctx), *input.ID); err != nil {
			return request, err
		}
	} else {
		if err := request.NewWithUser(currentUser); err != nil {
			return request, err
		}
	}

	if input.OrgID != nil {
		var org models.Organization
		err := org.FindByUUID(models.Tx(ctx), *input.OrgID)
		if err != nil {
			return models.Request{}, err
		}
		request.OrganizationID = org.ID
	}

	setStringField(input.Title, &request.Title)

	if input.NeededBefore == nil {
		request.NeededBefore = nulls.Time{}
	} else {
		neededBefore, err := domain.ConvertStringPtrToDate(input.NeededBefore)
		if err != nil {
			return models.Request{}, err
		}
		request.NeededBefore = nulls.NewTime(neededBefore)
	}

	setOptionalStringField(input.Description, &request.Description)

	if input.Size != nil {
		request.Size = *input.Size
	}

	setOptionalStringField(input.URL, &request.URL)
	setOptionalFloatField(input.Kilograms, &request.Kilograms)

	if input.Visibility == nil {
		request.Visibility = models.RequestVisibilitySame
	} else {
		request.Visibility = *input.Visibility
	}

	if input.PhotoID == nil {
		if request.ID > 0 {
			if err := request.RemoveFile(ctx); err != nil {
				return models.Request{}, err
			}
		}
	} else {
		if _, err := request.AttachPhoto(ctx, *input.PhotoID); err != nil {
			return models.Request{}, err
		}
	}

	if input.MeetingID == nil {
		request.MeetingID = nulls.Int{}
	} else {
		var meeting models.Meeting
		if err := meeting.FindByUUID(models.Tx(ctx), *input.MeetingID); err != nil {
			return models.Request{}, fmt.Errorf("invalid meetingID, %s", err)
		}
		request.MeetingID = nulls.NewInt(meeting.ID)
		request.DestinationID = meeting.LocationID
	}

	return request, nil
}

type requestInput struct {
	ID           *string
	OrgID        *string
	Title        *string
	Description  *string
	NeededBefore *string
	Destination  *LocationInput
	Origin       *LocationInput
	Size         *models.RequestSize
	URL          *string
	Kilograms    *float64
	PhotoID      *string
	MeetingID    *string
	Visibility   *models.RequestVisibility
}

// CreateRequest resolves the `createRequest` mutation.
func (r *mutationResolver) CreateRequest(ctx context.Context, input requestInput) (*models.Request, error) {
	cUser := models.CurrentUser(ctx)

	request, err := convertGqlRequestInputToDBRequest(ctx, input, cUser)
	if err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "CreateRequest.ProcessInput")
	}

	if !request.MeetingID.Valid {
		dest := convertLocation(*input.Destination)
		if err = dest.Create(ctx); err != nil {
			return &models.Request{}, domain.ReportError(ctx, err, "CreateRequest.SetDestination")
		}
		request.DestinationID = dest.ID
	}

	if err = request.Create(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "CreateRequest")
	}

	if input.Origin != nil {
		if err = request.SetOrigin(ctx, convertLocation(*input.Origin)); err != nil {
			return &models.Request{}, domain.ReportError(ctx, err, "CreateRequest.SetOrigin")
		}
	}

	return &request, nil
}

// UpdateRequest resolves the `updateRequest` mutation.
func (r *mutationResolver) UpdateRequest(ctx context.Context, input requestInput) (*models.Request, error) {
	cUser := models.CurrentUser(ctx)

	request, err := convertGqlRequestInputToDBRequest(ctx, input, cUser)
	if err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest.ProcessInput")
	}

	var dbRequest models.Request
	_ = dbRequest.FindByID(models.Tx(ctx), request.ID)
	if editable, err := dbRequest.IsEditable(ctx, cUser); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest.GetEditable")
	} else if !editable {
		return &models.Request{}, domain.ReportError(ctx, errors.New("attempt to update a non-editable request"),
			"UpdateRequest.NotEditable")
	}

	// TODO: move this to the end of the function? I did it back in 2019 but didn't make any not as to the reason
	if err := request.Update(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest")
	}

	if input.Destination != nil {
		if err := request.SetDestination(ctx, convertLocation(*input.Destination)); err != nil {
			return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest.SetDestination")
		}
	}

	if input.Origin == nil {
		if err := request.RemoveOrigin(ctx); err != nil {
			return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest.RemoveOrigin")
		}
	} else {
		if err := request.SetOrigin(ctx, convertLocation(*input.Origin)); err != nil {
			return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequest.SetOrigin")
		}
	}

	return &request, nil
}

// UpdateRequestStatus resolves the `updateRequestStatus` mutation.
func (r *mutationResolver) UpdateRequestStatus(ctx context.Context, input UpdateRequestStatusInput) (*models.Request, error) {
	var request models.Request
	if err := request.FindByUUID(models.Tx(ctx), input.ID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequestStatus.FindRequest")
	}

	cUser := models.CurrentUser(ctx)
	domain.NewExtra(ctx, "oldStatus", request.Status)
	domain.NewExtra(ctx, "newStatus", input.Status)

	if !cUser.CanUpdateRequestStatus(request, input.Status) {
		return &models.Request{}, domain.ReportError(ctx, errors.New("not allowed to change request status"),
			"UpdateRequestStatus.Unauthorized")
	}

	if err := request.SetProviderWithStatus(ctx, input.Status, input.ProviderUserID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error setting provider with status: "+err.Error()),
			"UpdateRequestStatus.SetProvider")
	}

	if err := request.Update(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "UpdateRequestStatus")
	}

	if err := request.DestroyPotentialProviders(ctx, input.Status, cUser); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error destroying request's potential providers: "+err.Error()),
			"UpdateRequestStatus.DestroyPotentialProviders")
	}

	return &request, nil
}

func (r *mutationResolver) AddMeAsPotentialProvider(ctx context.Context, requestID string) (*models.Request, error) {
	cUser := models.CurrentUser(ctx)
	var request models.Request
	if err := request.FindByUUIDForCurrentUser(ctx, requestID, cUser); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "AddMeAsPotentialProvider.FindRequest")
	}

	if request.Status != models.RequestStatusOpen {
		return &models.Request{}, domain.ReportError(ctx, errors.New(
			"Can only create PotentialProvider for a Request that has Status=Open. Got "+request.Status.String()),
			"AddMeAsPotentialProvider.BadRequestStatus")
	}

	var provider models.PotentialProvider
	if err := provider.NewWithRequestUUID(ctx, requestID, cUser.ID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error preparing potential provider: "+err.Error()),
			"AddMeAsPotentialProvider")
	}

	if err := provider.Create(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error creating potential provider: "+err.Error()),
			"AddMeAsPotentialProvider")
	}

	return &request, nil
}

func (r *mutationResolver) RemoveMeAsPotentialProvider(ctx context.Context, requestID string) (*models.Request, error) {
	cUser := models.CurrentUser(ctx)

	var provider models.PotentialProvider

	if err := provider.FindWithRequestUUIDAndUserUUID(ctx, requestID, cUser.UUID.String(), cUser); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("unable to find PotentialProvider in order to delete it: "+err.Error()),
			"RemoveMeAsPotentialProvider")
	}

	var request models.Request
	if err := request.FindByUUID(models.Tx(ctx), requestID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "RemoveMeAsPotentialProvider.FindRequest")
	}

	domain.NewExtra(ctx, "request", request.UUID)

	if err := provider.Destroy(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error removing potential provider: "+err.Error()),
			"RemoveMeAsPotentialProvider")
	}

	if err := request.FindByUUID(models.Tx(ctx), requestID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "RemoveMeAsPotentialProvider.FindRequest")
	}

	return &request, nil
}

func (r *mutationResolver) RejectPotentialProvider(ctx context.Context, requestID, userID string) (*models.Request, error) {
	cUser := models.CurrentUser(ctx)
	var provider models.PotentialProvider
	if err := provider.FindWithRequestUUIDAndUserUUID(ctx, requestID, userID, cUser); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("unable to find PotentialProvider in order to delete it: "+err.Error()),
			"RemovePotentialProvider")
	}

	var request models.Request
	if err := request.FindByUUID(models.Tx(ctx), requestID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "RemovePotentialProvider.FindRequest")
	}

	domain.NewExtra(ctx, "request", request.UUID)

	if err := provider.Destroy(ctx); err != nil {
		return &models.Request{}, domain.ReportError(ctx, errors.New("error removing potential provider: "+err.Error()),
			"RemovePotentialProvider")
	}

	if err := request.FindByUUID(models.Tx(ctx), requestID); err != nil {
		return &models.Request{}, domain.ReportError(ctx, err, "RemovePotentialProvider.FindRequest")
	}

	return &request, nil
}

func (r *mutationResolver) MarkRequestAsDelivered(ctx context.Context, requestID string) (*models.Request, error) {
	input := UpdateRequestStatusInput{Status: models.RequestStatusDelivered, ID: requestID}

	return r.UpdateRequestStatus(ctx, input)
}

func (r *mutationResolver) MarkRequestAsReceived(ctx context.Context, requestID string) (*models.Request, error) {
	input := UpdateRequestStatusInput{Status: models.RequestStatusCompleted, ID: requestID}

	return r.UpdateRequestStatus(ctx, input)
}
