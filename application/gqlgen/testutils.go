package gqlgen

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/ast"
)

func testContext(sel ast.SelectionSet) context.Context {

	ctx := context.Background()

	rqCtx := &graphql.RequestContext{}
	ctx = graphql.WithRequestContext(ctx, rqCtx)

	root := &graphql.ResolverContext{
		Field: graphql.CollectedField{
			Selections: sel,
		},
	}
	ctx = graphql.WithResolverContext(ctx, root)

	return ctx
}
