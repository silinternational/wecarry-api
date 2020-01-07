package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	i18n "github.com/gobuffalo/mw-i18n"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/listeners"
)

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
			Env: domain.Env.GoEnv,
			PreWares: []buffalo.PreWare{
				cors.New(cors.Options{
					AllowCredentials: true,
					AllowedOrigins:   []string{domain.Env.UIURL},
					AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
					AllowedHeaders:   []string{"*"},
				}).Handler,
			},
			SessionName:  "_wecarry_session",
			SessionStore: sessions.NewCookieStore([]byte(domain.Env.SessionSecret)),
		})

		// If you add a new status entry here, then also add it to getErrorCodeFromStatus
		app.ErrorHandlers[http.StatusUnauthorized] = customErrorHandler        // 401
		app.ErrorHandlers[http.StatusNotFound] = customErrorHandler            // 404
		app.ErrorHandlers[http.StatusMethodNotAllowed] = customErrorHandler    // 405
		app.ErrorHandlers[http.StatusInternalServerError] = customErrorHandler // 500

		// Initialize and attach "rollbar" to context
		app.Use(domain.RollbarMiddleware)

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		//  Added for authorization
		app.Use(setCurrentUser)
		app.Middleware.Skip(setCurrentUser, statusHandler)

		var err error
		domain.T, err = i18n.New(packr.New("locales", "../locales"), "en")
		if err != nil {
			_ = app.Stop(err)
		}
		app.Use(domain.T.Middleware())

		app.GET("/site/status", statusHandler)
		app.Middleware.Skip(buffalo.RequestLogger, statusHandler)

		app.POST("/gql/", gqlHandler)

		app.POST("/upload/", uploadHandler)

		auth := app.Group("/auth")
		auth.Middleware.Skip(setCurrentUser, authRequest, authCallback, authDestroy)

		auth.POST("/login", authRequest)

		auth.GET("/callback", authCallback)  // for Google Oauth
		auth.POST("/callback", authCallback) // for SAML

		auth.GET("/logout", authDestroy)

		listeners.RegisterListeners()
	}

	return app
}
