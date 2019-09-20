package grifts

import (
	"github.com/silinternational/wecarry-api/actions"

	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
