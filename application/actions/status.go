package actions

import "github.com/gobuffalo/buffalo"

// statusHandler is a handler to respond with the current site status
func statusHandler(c buffalo.Context) error {
	return c.Render(200, r.JSON(map[string]string{"status": "good"}))
}
