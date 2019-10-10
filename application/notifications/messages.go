package notifications

import "github.com/silinternational/wecarry-api/models"

type Message struct {
	Template string
	Data     map[string]interface{}
	From     models.User
	To       models.User
}
