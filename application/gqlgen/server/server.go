package main

import (
	//"log"
	//"net/http"
	//"os"
	//
	//"context"
	//
	//"fmt"
	//
	//"github.com/99designs/gqlgen/graphql"
	//"github.com/99designs/gqlgen/handler"
	//"github.com/silinternational/handcarry-api/gqlgen"
	//"github.com/go-chi/chi"
)

const defaultPort = "8080"

//func checkIfAuthZd(ctx context.Context, role gqlgen.Role) error {
//	user, err := gqlgen.GetGQLUserFromContext(ctx)
//	if err != nil {
//		return err
//	}
//
//	if !user.IsAuthZd(role.String()) {
//		// block calling the next resolver
//		return fmt.Errorf("Access denied")
//	}
//
//	return nil
//}
//
func main() {
	
}
//	port := os.Getenv("PORT")
//	if port == "" {
//		port = defaultPort
//	}
//
//	cfg := gqlgen.Config{Resolvers: &gqlgen.Resolver{}}
//
//	cfg.Directives.IsAuthZd = func(ctx context.Context, _ interface{}, next graphql.Resolver, role gqlgen.Role) (interface{}, error) {
//		err := checkIfAuthZd(ctx, role)
//		if err != nil {
//			return nil, err
//		}
//
//		// or let it pass through
//		return next(ctx)
//	}
//
//	router := chi.NewRouter()
//	router.Use(gqlgen.AuthNMiddleware())
//
//	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
//	router.Handle("/query", handler.GraphQL(gqlgen.NewExecutableSchema(cfg)))
//
//	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
//
//	err := http.ListenAndServe(":"+port, router)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
