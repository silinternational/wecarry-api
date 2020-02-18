package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestMeetingInvitation_Meeting(t *testing.T) {
	type fields struct {
		ID        int
		CreatedAt time.Time
		UpdatedAt time.Time
		MeetingID int
		InviterID int
		Secret    uuid.UUID
		Email     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    Meeting
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MeetingInvitation{
				ID:        tt.fields.ID,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				MeetingID: tt.fields.MeetingID,
				InviterID: tt.fields.InviterID,
				Secret:    tt.fields.Secret,
				Email:     tt.fields.Email,
			}
			got, err := m.Meeting()
			if (err != nil) != tt.wantErr {
				t.Errorf("Meeting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Meeting() got = %v, want %v", got, tt.want)
			}
		})
	}
}
