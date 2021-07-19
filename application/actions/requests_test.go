package actions

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type RequestTestFixtures struct {
	models.Organization
	models.Users
	models.Requests
	models.Threads
}

func (as *ActionSuite) Test_requestsGet() {
	f := createFixturesForRequests(as)

	tests := []struct {
		name       string
		user       models.User
		requestID  string
		wantStatus int
	}{
		{
			name:       "authn error",
			user:       models.User{},
			requestID:  f.Requests[0].UUID.String(),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "authz error",
			user:       f.Users[1],
			requestID:  f.Requests[3].UUID.String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad request ID",
			user:       f.Users[1],
			requestID:  "1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			user:       f.Users[1],
			requestID:  domain.GetUUID().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no error",
			user:       f.Users[1],
			requestID:  f.Requests[0].UUID.String(),
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/requests/" + tt.requestID)
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Get()

			body := res.Body.String()
			as.Equal(tt.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tt.wantStatus == http.StatusOK {
				as.Contains(body, fmt.Sprintf(`"id":"%s"`, tt.requestID))
			}
		})
	}
}

func (as *ActionSuite) Test_requestsCreate() {
	f := createFixturesForRequests(as)

	destination := api.Location{
		Description: "location description",
		Country:     "XX",
		Latitude:    1.1,
		Longitude:   2.2,
	}

	goodRequest := api.RequestCreateInput{
		Destination:    destination,
		OrganizationID: f.Organization.UUID,
		Size:           api.RequestSize(models.RequestSizeSmall),
		Title:          "request title",
	}
	badRequest := api.RequestCreateInput{}

	tests := []struct {
		name       string
		user       models.User
		request    api.RequestCreateInput
		wantStatus int
	}{
		{
			name:       "authn error",
			user:       models.User{},
			request:    goodRequest,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad input",
			user:       f.Users[1],
			request:    badRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "good input",
			user:       f.Users[1],
			request:    goodRequest,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/requests")
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Post(&tt.request)

			body := res.Body.String()
			as.Equal(tt.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tt.wantStatus == http.StatusOK {
				wantData := []string{
					`"title":"` + tt.request.Title,
					`"created_by":{"id":"` + tt.user.UUID.String(),
					`"organization":{"id":"` + f.Organization.UUID.String(),
					`"destination":{"description":"` + destination.Description,
				}
				as.verifyResponseData(wantData, body, "")
			}
		})
	}
}

func (as *ActionSuite) Test_convertRequest() {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	creator := userFixtures.Users[0]
	provider := userFixtures.Users[1]

	requestFixtures := test.CreateRequestFixtures(as.DB, 2, false, creator.ID)

	meeting := test.CreateMeetingFixtures(as.DB, 1, creator)[0]

	ctx := test.CtxWithUser(creator)

	min := requestFixtures[0]
	min.ProviderID = nulls.Int{}
	min.Description = nulls.String{}
	min.OriginID = nulls.Int{}
	min.NeededBefore = nulls.Time{}
	min.Kilograms = nulls.Float64{}
	min.URL = nulls.String{}
	min.FileID = nulls.Int{}
	min.MeetingID = nulls.Int{}

	// because Pop doesn't update child objects when the ID changes to 0
	min.Provider = models.User{}
	min.Origin = models.Location{}
	min.PhotoFile = models.File{}
	min.Meeting = models.Meeting{}

	as.NoError(as.DB.Save(&min))

	full := requestFixtures[1]
	full.ProviderID = nulls.NewInt(provider.ID)
	full.FileID = nulls.NewInt(test.CreateFileFixture(as.DB).ID)
	full.MeetingID = nulls.NewInt(meeting.ID)

	as.NoError(as.DB.Save(&full))

	tests := []struct {
		name    string
		request models.Request
	}{
		{
			name:    "minimal",
			request: min,
		},
		{
			name:    "full",
			request: full,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			apiRequest, err := models.ConvertRequest(ctx, tt.request)
			as.NoError(err)

			as.NoError(as.DB.Load(&tt.request))
			as.verifyApiRequest(ctx, tt.request, apiRequest, "api.Request is not correct")
		})
	}
}

func (as *ActionSuite) verifyApiRequest(ctx context.Context, request models.Request, apiRequest api.Request, msg string) {
	as.Equal(request.UUID.String(), apiRequest.ID.String(), msg+", ID is not correct")

	isEditable, err := request.IsEditable(as.DB, models.CurrentUser(ctx))
	as.NoError(err)
	as.Equal(isEditable, apiRequest.IsEditable, msg+", IsEditable is not correct")

	as.Equal(string(request.Status), string(apiRequest.Status), msg+", Status is not correct")

	as.verifyUser(request.CreatedBy, apiRequest.CreatedBy, msg+", CreatedBy is not correct")

	if request.ProviderID.Valid {
		as.NotNil(apiRequest.Provider, msg+", Provider is null but should not be")
		as.verifyUser(request.Provider, *apiRequest.Provider, msg+", Provider is not correct")

	} else {
		as.Nil(apiRequest.Provider, msg+", Provider should be null but is not")
	}

	potentialProviders, err := request.GetPotentialProviders(as.DB, models.CurrentUser(ctx))
	as.NoError(err)
	as.verifyPotentialProviders(potentialProviders, apiRequest.PotentialProviders, msg+", potential providers are not correct")

	as.verifyOrganization(request.Organization, apiRequest.Organization, msg+", Organization is not correct")

	as.Equal(string(request.Visibility), string(apiRequest.Visibility), msg+", Visibility is not correct")

	as.Equal(request.Title, apiRequest.Title, msg+", Title is not correct")

	as.Equal(request.Description, apiRequest.Description, msg+", Description is not correct")

	as.verifyLocation(request.Destination, apiRequest.Destination, msg+", Destination is not correct")

	if request.OriginID.Valid {
		as.NotNil(apiRequest.Origin, msg+", Origin is null but should not be")
		as.verifyLocation(request.Origin, *apiRequest.Origin, msg+", Origin is not correct")
	} else {
		as.Nil(apiRequest.Origin, msg+", Origin should be null but is not")
	}

	as.Equal(string(request.Size), string(apiRequest.Size), msg+", Size is not correct")

	as.True(request.CreatedAt.Equal(apiRequest.CreatedAt), msg+", CreatedAt is not correct")

	as.True(request.UpdatedAt.Equal(apiRequest.UpdatedAt), msg+", UpdatedAt is not correct")

	if request.NeededBefore.Valid {
		as.NotNil(apiRequest.NeededBefore, msg+", NeededBefore is null but should not be")
		as.Equal(request.NeededBefore.Time.Format(domain.DateFormat),
			apiRequest.NeededBefore.String, msg+", NeededBefore is not correct")
	} else {
		as.False(apiRequest.NeededBefore.Valid, msg+", NeededBefore should be null but is not")
	}

	if request.Kilograms.Valid {
		as.NotNil(apiRequest.Kilograms, msg+", Kilograms is null but should not be")
		as.Equal(request.Kilograms, apiRequest.Kilograms, msg+", Kilograms is not correct")
	} else {
		as.False(apiRequest.Kilograms.Valid, msg+", Kilograms should be null but is not")
	}

	if request.URL.Valid {
		as.NotNil(apiRequest.URL, msg+", URL is null but should not be")
		as.Equal(request.URL, apiRequest.URL, msg+", URL is not correct")
	} else {
		as.False(apiRequest.URL.Valid, msg+", URL should be null but is not")
	}

	if request.FileID.Valid {
		as.NotNil(apiRequest.Photo, msg+", Photo is null but should not be")
		as.verifyFile(request.PhotoFile, *apiRequest.Photo, msg+", Photo is not correct")
	} else {
		as.Nil(apiRequest.Photo, msg+", Photo should be null but is not")
	}

	if request.MeetingID.Valid {
		as.NotNil(apiRequest.Meeting, msg+", Meeting is null but should not be")
		as.verifyMeeting(request.Meeting, *apiRequest.Meeting, msg+", Meeting is not correct")
	} else {
		as.Nil(apiRequest.Meeting, msg+", Meeting should be null but is not")
	}
}

func (as *ActionSuite) Test_convertCreateRequestInput() {
	userFixtures := test.CreateUserFixtures(as.DB, 1)
	creator := userFixtures.Users[0]

	ctx := test.CtxWithUser(creator)

	destination := api.Location{
		Description: "destination",
		Country:     "XX",
		Latitude:    1.1,
		Longitude:   2.2,
	}
	origin := api.Location{
		Description: "origin",
		Country:     "ZZ",
		Latitude:    -1.1,
		Longitude:   -2.2}
	file := test.CreateFileFixture(as.DB)

	min := api.RequestCreateInput{
		Destination:    destination,
		OrganizationID: userFixtures.Organization.UUID,
		Size:           api.RequestSize(models.RequestSizeSmall),
		Title:          "request title 1",
	}

	full := api.RequestCreateInput{
		Description:    nulls.NewString("request description"),
		Destination:    destination,
		Kilograms:      nulls.NewFloat64(1.0),
		NeededBefore:   nulls.NewString(time.Now().Format(domain.DateFormat)),
		Origin:         &origin,
		OrganizationID: userFixtures.Organization.UUID,
		PhotoID:        nulls.NewUUID(file.UUID),
		Size:           api.RequestSize(models.RequestSizeMedium),
		Title:          "request title 2",
		Visibility:     api.RequestVisibility(models.RequestVisibilityAll),
	}

	tests := []struct {
		name  string
		input api.RequestCreateInput
	}{
		{
			name:  "minimal",
			input: min,
		},
		{
			name:  "full",
			input: full,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			apiRequest, err := convertRequestCreateInput(ctx, tt.input)
			as.NoError(err)

			as.NoError(as.DB.Load(&tt.input))
			as.verifyRequestCreateInput(ctx, tt.input, apiRequest)
		})
	}
}

func (as *ActionSuite) verifyRequestCreateInput(ctx context.Context, input api.RequestCreateInput, request models.Request) {
	as.Equal(models.CurrentUser(ctx).ID, request.CreatedByID, "CreatedBy is not correct")

	as.Equal(input.Description, request.Description, "Description is not correct")

	var destination models.Location
	as.NoError(as.DB.Find(&destination, request.DestinationID))
	as.verifyLocation(destination, input.Destination, "Destination is not correct")

	if input.Kilograms.Valid {
		as.NotNil(request.Kilograms, "Kilograms is null but should not be")
		as.Equal(input.Kilograms, request.Kilograms, "Kilograms is not correct")
	} else {
		as.False(request.Kilograms.Valid, "Kilograms should be null but is not")
	}

	if input.NeededBefore.Valid {
		as.NotNil(request.NeededBefore, "NeededBefore is null but should not be")
		as.Equal(input.NeededBefore.String,
			request.NeededBefore.Time.Format(domain.DateFormat), "NeededBefore is not correct")
	} else {
		as.False(request.NeededBefore.Valid, "NeededBefore should be null but is not")
	}

	if input.Origin != nil {
		as.NotNil(request.Origin, "Origin is null but should not be")
		var origin models.Location
		as.NoError(as.DB.Find(&origin, request.OriginID))
		as.verifyLocation(origin, *input.Origin, "Origin is not correct")
	} else {
		as.False(request.OriginID.Valid, "Origin should be null but is not")
	}

	var organization models.Organization
	as.NoError(as.DB.Find(&organization, request.OrganizationID))
	as.Equal(input.OrganizationID, organization.UUID, "Organization is not correct")

	if input.PhotoID.Valid {
		as.NotNil(request.FileID, "Photo is null but should not be")
		var file models.File
		as.NoError(as.DB.Find(&file, request.FileID))
		as.Equal(input.PhotoID.UUID, file.UUID, "Photo is not correct")
	} else {
		as.False(request.FileID.Valid, "Photo should be null but is not")
	}

	as.Equal(string(input.Size), string(request.Size), "Size is not correct")

	as.Equal(input.Title, request.Title, "Title is not correct")

	as.Equal(string(input.Visibility), string(request.Visibility), "Visibility is not correct")
}

func (as *ActionSuite) Test_requestsUpdate() {
	f := createFixturesForRequests(as)

	goodRequest := api.RequestUpdateInput{
		Description: nulls.NewString("new description"),
	}

	badRequestField := api.RequestCreateInput{
		OrganizationID: f.Organization.UUID,
	}

	empty := ""
	badRequestData := api.RequestUpdateInput{
		Title: &empty,
	}

	tests := []struct {
		name       string
		user       models.User
		input      interface{}
		request    models.Request
		wantStatus int
	}{
		{
			name:       "authn error",
			user:       models.User{},
			input:      goodRequest,
			request:    f.Requests[0],
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad input field",
			user:       f.Users[1],
			input:      badRequestField,
			request:    f.Requests[1],
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad input data",
			user:       f.Users[1],
			input:      badRequestData,
			request:    f.Requests[1],
			wantStatus: http.StatusInternalServerError, // TODO: this needs to be StatusBadRequest
		},
		{
			name:       "good input",
			user:       f.Users[0],
			input:      goodRequest,
			request:    f.Requests[2],
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			req := as.JSON("/requests/" + tt.request.UUID.String())
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tt.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Put(&tt.input)

			body := res.Body.String()
			as.Equal(tt.wantStatus, res.Code, "incorrect status code returned, body: %s", body)

			if tt.wantStatus == http.StatusOK {
				wantData := []string{
					`"created_by":{"id":"` + tt.user.UUID.String(),
					`"organization":{"id":"` + f.Organization.UUID.String(),
					`"description":"` + goodRequest.Description.String,
				}
				as.verifyResponseData(wantData, body, "")
			}
		})
	}
}

func (as *ActionSuite) Test_convertUpdateRequestInput() {
	userFixtures := test.CreateUserFixtures(as.DB, 1)
	requests := test.CreateRequestFixtures(as.DB, 2, false, userFixtures.Users[0].ID)

	creator := userFixtures.Users[0]

	ctx := test.CtxWithUser(creator)

	min := api.RequestUpdateInput{}

	destination := api.Location{
		Description: "destination",
		Country:     "XX",
		Latitude:    1.1,
		Longitude:   2.2,
	}
	file := test.CreateFileFixture(as.DB)
	origin := api.Location{
		Description: "origin",
		Country:     "ZZ",
		Latitude:    -1.1,
		Longitude:   -2.2,
	}
	size := api.RequestSize(models.RequestSizeMedium)
	title := "request title"
	visibility := api.RequestVisibility(models.RequestVisibilityAll)
	full := api.RequestUpdateInput{
		Description:  nulls.NewString("request description"),
		Destination:  &destination,
		Kilograms:    nulls.NewFloat64(1.0),
		NeededBefore: nulls.NewString(time.Now().Format(domain.DateFormat)),
		Origin:       &origin,
		PhotoID:      nulls.NewUUID(file.UUID),
		Size:         &size,
		Title:        &title,
		Visibility:   &visibility,
	}

	tests := []struct {
		name    string
		input   api.RequestUpdateInput
		request models.Request
	}{
		{
			name:    "minimal",
			input:   min,
			request: requests[0],
		},
		{
			name:    "full",
			input:   full,
			request: requests[1],
		},
	}
	for _, tt := range tests {
		as.T().Run(tt.name, func(t *testing.T) {
			apiRequest, err := convertRequestUpdateInput(ctx, tt.input, tt.request.UUID.String())
			as.NoError(err)

			as.NoError(as.DB.Load(&tt.input))
			as.verifyRequestUpdateInput(ctx, tt.input, tt.request, apiRequest)
		})
	}
}

func (as *ActionSuite) verifyRequestUpdateInput(ctx context.Context, input api.RequestUpdateInput, oldRequest, newRequest models.Request) {
	as.Equal(models.CurrentUser(ctx).ID, newRequest.CreatedByID, "CreatedBy is not correct")

	as.Equal(input.Description, newRequest.Description, "Description is not correct")

	var destination models.Location
	as.NoError(as.DB.Find(&destination, newRequest.DestinationID))
	msg := "Destination is not correct"
	if input.Destination != nil {
		as.verifyLocation(destination, *input.Destination, msg)
	} else {
		as.Equal(destination.Description, oldRequest.Destination.Description, msg+", Description is not correct")
		as.Equal(destination.Country, oldRequest.Destination.Country, msg+", Country is not correct")
	}

	if input.Kilograms.Valid {
		as.NotNil(newRequest.Kilograms, "Kilograms is null but should not be")
		as.Equal(input.Kilograms, newRequest.Kilograms, "Kilograms is not correct")
	} else {
		as.False(newRequest.Kilograms.Valid, "Kilograms should be null but is not")
	}

	if input.NeededBefore.Valid {
		as.NotNil(newRequest.NeededBefore, "NeededBefore is null but should not be")
		as.Equal(input.NeededBefore.String,
			newRequest.NeededBefore.Time.Format(domain.DateFormat), "NeededBefore is not correct")
	} else {
		as.False(newRequest.NeededBefore.Valid, "NeededBefore should be null but is not")
	}

	if input.Origin != nil {
		as.NotNil(newRequest.Origin, "Origin is null but should not be")
		var origin models.Location
		as.NoError(as.DB.Find(&origin, newRequest.OriginID))
		as.verifyLocation(origin, *input.Origin, "Origin is not correct")
	} else {
		as.False(newRequest.OriginID.Valid, "Origin should be null but is not")
	}

	if input.PhotoID.Valid {
		as.NotNil(newRequest.FileID, "Photo is null but should not be")
		var file models.File
		as.NoError(as.DB.Find(&file, newRequest.FileID))
		as.Equal(input.PhotoID.UUID, file.UUID, "Photo is not correct")
	} else {
		as.False(newRequest.FileID.Valid, "Photo should be null but is not")
	}

	msg = "Size is not correct"
	if input.Size != nil {
		as.Equal(string(*input.Size), string(newRequest.Size), msg)
	} else {
		as.Equal(string(oldRequest.Size), string(newRequest.Size), msg)
	}

	msg = "Title is not correct"
	if input.Title != nil {
		as.Equal(*input.Title, newRequest.Title, msg)
	} else {
		as.Equal(oldRequest.Title, newRequest.Title, msg)
	}

	msg = "Visibility is not correct"
	if input.Visibility != nil {
		as.Equal(string(*input.Visibility), string(newRequest.Visibility), msg)
	} else {
		as.Equal(string(oldRequest.Visibility), string(newRequest.Visibility), msg)
	}
}

func (as *ActionSuite) Test_requestsUpdateStatus() {
	f := createFixturesForUpdateRequestStatus(as)

	request := f.Requests[0]
	creator := f.Users[0]
	provider := f.Users[1]
	providerUUID := provider.UUID.String()

	steps := []struct {
		name       string
		status     models.RequestStatus
		user       models.User
		requestID  string
		wantStatus int
		wantKey    api.ErrorKey
		providerID *string
	}{
		{
			name:       "non-creator can't change status to ACCEPTED",
			status:     models.RequestStatusAccepted,
			user:       provider,
			requestID:  request.UUID.String(),
			wantStatus: http.StatusBadRequest,
			wantKey:    api.ErrorUpdateRequestStatusBadStatus,
			providerID: &providerUUID,
		},
		{
			name:       "request ID not found",
			status:     models.RequestStatusAccepted,
			user:       creator,
			requestID:  domain.GetUUID().String(),
			wantStatus: http.StatusNotFound,
			wantKey:    api.ErrorUpdateRequestStatusNotFound,
			providerID: &providerUUID,
		},
		{
			name:       "null provider ID",
			status:     models.RequestStatusAccepted,
			user:       creator,
			requestID:  request.UUID.String(),
			wantStatus: http.StatusBadRequest,
			wantKey:    api.ErrorUpdateRequestStatusBadProvider,
			providerID: nil,
		},
		{
			name:       "creator can change status to ACCEPTED",
			status:     models.RequestStatusAccepted,
			user:       creator,
			requestID:  request.UUID.String(),
			wantStatus: http.StatusOK,
			providerID: &providerUUID,
		},
	}

	for _, step := range steps {
		input := api.RequestUpdateStatusInput{
			Status:         api.RequestStatus(step.status),
			ProviderUserID: step.providerID,
		}

		req := as.JSON("/requests/" + step.requestID + "/status")
		req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", step.user.Nickname)
		req.Headers["content-type"] = "application/json"
		res := req.Put(&input)

		body := res.Body.String()
		as.Equal(step.wantStatus, res.Code, `step "%s", incorrect status code returned, body: %s`, step.name, body)

		if step.wantStatus != http.StatusOK {
			as.verifyResponseData([]string{string(step.wantKey)}, body, fmt.Sprintf(`step "%s", `, step.name))
			continue
		}
		wantData := []string{
			`"status":"` + step.status.String(),
			`"provider":{"id":"` + providerUUID,
		}

		as.verifyResponseData(wantData, body, fmt.Sprintf(`step "%s", `, step.name))
	}
}
