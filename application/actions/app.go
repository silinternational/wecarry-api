package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/pop/popmw"
	"github.com/gobuffalo/envy"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
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
			Env: ENV,
			PreWares: []buffalo.PreWare{
				cors.New(cors.Options{
					AllowCredentials: true,
					AllowedOrigins:   []string{envy.Get(domain.UIURLEnv, "*")},
					AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
					AllowedHeaders:   []string{"*"},
				}).Handler,
			},
			SessionName:  "_wecarry_session",
			SessionStore: sessions.NewCookieStore([]byte(envy.Get("SESSION_SECRET", "testing"))),
		})

		// Initialize and attach "rollbar" to context
		app.Use(domain.RollbarMiddleware)

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		//  Added for authorization
		app.Use(SetCurrentUser)
		app.Middleware.Skip(SetCurrentUser, HomeHandler)

		app.GET("/", HomeHandler)
		app.POST("/gql/", GQLHandler)

		app.POST("/upload/", UploadHandler)

		auth := app.Group("/auth")
		auth.Middleware.Skip(SetCurrentUser, AuthRequest, AuthCallback, AuthDestroy)

		auth.POST("/login", AuthRequest)

		auth.GET("/callback", AuthCallback)  // for Google Oauth
		auth.POST("/callback", AuthCallback) // for SAML

		auth.GET("/logout", AuthDestroy)
	}

	return app
}
