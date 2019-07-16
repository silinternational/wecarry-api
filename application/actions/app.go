package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	csrf "github.com/gobuffalo/mw-csrf"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/markbates/goth/gothic"

	"github.com/silinternational/handcarry-api/models"

	"github.com/gobuffalo/buffalo-pop/pop/popmw"
	contenttype "github.com/gobuffalo/mw-contenttype"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_application_session",
		})

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Set the request content type to JSON
		app.Use(contenttype.Set("application/json"))

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		//  Added for authorization
		app.Use(SetCurrentUser)
		app.Middleware.Skip(SetCurrentUser, HomeHandler, AuthLogin, AuthCallback)

		app.GET("/", HomeHandler)

		auth := app.Group("/auth")
		auth.GET("/login", AuthLogin)
		auth.Middleware.Skip(SetCurrentUser, AuthLogin, AuthCallback)
		bah := buffalo.WrapHandlerFunc(gothic.BeginAuthHandler)
		auth.GET("/logout", AuthDestroy)
		auth.GET("/{provider}", bah)
		auth.GET("/{provider}/callback", AuthCallback)  // "GET" for Google
		auth.POST("/{provider}/callback", AuthCallback) //  "POST" for saml2
		auth.Middleware.Skip(csrf.New, AuthCallback)    //  Don't require csrf on auth

	}

	return app
}
