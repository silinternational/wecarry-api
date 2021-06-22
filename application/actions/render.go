package actions

import (
	"encoding/json"
	"io"

	"github.com/gobuffalo/buffalo/render"
	"github.com/liip/sheriff"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		DefaultContentType: "application/json",
	})
}

type sheriffRenderer struct {
	value  interface{}
	groups []string
}

func (s sheriffRenderer) ContentType() string {
	return "application/json; charset=utf-8"
}

func (s sheriffRenderer) Render(w io.Writer, d render.Data) error {
	o := &sheriff.Options{
		Groups: s.groups,
	}

	data, err := sheriff.Marshal(o, s.value)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(data)
}
