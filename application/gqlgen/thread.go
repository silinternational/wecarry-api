package gqlgen

import (
	"github.com/silinternational/handcarry-api/domain"
)

func ThreadSimpleFields() map[string]string {
	return map[string]string{
		"id":        "uuid",
		"postID":    "post_id",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
}

func getSelectFieldsForThreads(requestFields []string) []string {
	selectFields := GetSelectFieldsFromRequestFields(ThreadSimpleFields(), requestFields)

	// Ensure we can get participants via the thread ID
	if domain.IsStringInSlice(ParticipantsField, requestFields) {
		selectFields = append(selectFields, "id")
	}

	// Ensure we can get the post via the post ID
	if domain.IsStringInSlice(PostField, requestFields) {
		selectFields = append(selectFields, "post_id")
	}

	return selectFields
}
