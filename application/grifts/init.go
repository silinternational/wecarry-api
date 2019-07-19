package grifts

import (
	"github.com/silinternational/handcarry-api/actions"

	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
