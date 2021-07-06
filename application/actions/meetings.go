package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertMeeting(meeting models.Meeting) api.Meeting {
	return api.Meeting{
		ID:          meeting.UUID,
		Name:        meeting.Name,
		Description: meeting.Description.String,
		StartDate:   meeting.StartDate,
		EndDate:     meeting.EndDate,
		CreatedAt:   meeting.CreatedAt,
		UpdatedAt:   meeting.UpdatedAt,
		MoreInfoURL: meeting.MoreInfoURL.String,
	}
}
