package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
)

const defaultPort = "3000"

func main() {
	port := domain.Env.PlaygroundPort
	if port == "" {
		port = defaultPort
	}

	cfg := gqlgen.Config{Resolvers: &gqlgen.Resolver{}}

	router := chi.NewRouter()
	// router.Use(gqlgen.AuthNMiddleware())

	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	router.Handle("/query", handler.GraphQL(gqlgen.NewExecutableSchema(cfg)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal(err)
	}
}
