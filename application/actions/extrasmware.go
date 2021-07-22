package actions

import (
	"github.com/gobuffalo/buffalo"
)

func setHTTPExtras(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		newExtra(c, "request_method", c.Request().Method)
		newExtra(c, "request_url", c.Request().URL.Path)

		err := next(c)
		// do some work after calling the next handler
		return err
	}
}
