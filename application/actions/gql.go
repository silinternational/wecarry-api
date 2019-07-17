package actions

import (
	"github.com/gobuffalo/buffalo"
)

func GQLHandler(c buffalo.Context) error {


	return c.Render(200, r.JSON(map[string]string{"message": "Welcome to GQL!"}))
}


//func GQLHandler(c buffalo.Context) error {
//
//
//	cfg := gqlgen.Config{Resolvers: &gqlgen.Resolver{}}
//	executable := gqlgen.NewExecutableSchema(cfg)
//
//	response := executable.Query()
//
//
//
//
//
//	return c.Render(200, r.JSON(map[string]string{"message": "Welcome to GQL!"}))
//}
