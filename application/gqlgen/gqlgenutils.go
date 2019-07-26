package gqlgen

import (
	"github.com/gobuffalo/nulls"
)

const MessagesField = "messages"
const ParticipantsField = "participants"
const PostField = "post"
const PostIDField = "postID"
const SenderField = "sender"
const CreatedByField = "createdBy"

func GetStringFromNullsString(inString nulls.String) *string {
	output := ""
	if inString.Valid {
		output = inString.String
	}

	return &output
}
