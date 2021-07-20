package actions

import (
	"fmt"
	"net/http"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
)

func WatchesMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		method := c.Request().Method

		switch method {
		case http.MethodDelete:
			if err := initWatchesDelete(c); err != nil {
				return err
			}
		}

		err := next(c)
		// do some work after calling the next handler
		return err
	}
}

func initWatchesDelete(c buffalo.Context) error {
	id, err := getUUIDFromParam(c, watchIDKey)
	if err != nil {
		return reportError(c, err)
	}
	domain.NewExtra(c, watchIDKey, id)
	c.Set(watchIDKey, id)
	fmt.Printf("\nInit Watches Delete, url: %+v\n", c.Request().URL)
	return nil
}

func getWatchIDFromContext(c buffalo.Context) string {
	id := c.Value(watchIDKey)
	return fmt.Sprintf("%v", id)
}
