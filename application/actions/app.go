package actions

// WeCarry API
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: https
//     Host: localhost
//     BasePath: /
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - oauth2:
//
//     SecurityDefinitions:
//     bearerAuth:
//         type: http
//         scheme: bearer
//
// swagger:meta

import (
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

		var err error
		domain.T, err = i18n.New(packr.New("locales", "../locales"), "en")
		if err != nil {
			_ = app.Stop(err)
		}
		app.Use(domain.T.Middleware())

		registerCustomErrorHandler(app)

		// Initialize and attach "rollbar" to context
		app.Use(domain.RollbarMiddleware)

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		//  Added for authorization
		app.Use(setCurrentUser)
		app.Middleware.Skip(setCurrentUser, statusHandler, serviceHandler)

		app.GET("/site/status", statusHandler)
		app.Middleware.Skip(buffalo.RequestLogger, statusHandler)

		app.POST("/gql/", gqlHandler)

		eventsGroup := app.Group("/events")
		eventsGroup.GET("/", meetingsList)
		eventsGroup.POST("/", meetingsJoin)

		app.POST("/messages/", messagesCreate)

		threadsGroup := app.Group("/threads")
		threadsGroup.GET("/", threadsMine)
		threadsGroup.PUT("/{thread_id}/read", threadsMarkAsRead)

		requestsGroup := app.Group("/requests")
		requestsGroup.GET("/", requestsList)
		requestsGroup.POST("/", requestsCreate)
		requestsGroup.GET("/{request_id}", requestsGet)

		requestsGroup.POST("/{request_id}/potentialprovider", requestsAddMeAsPotentialProvider)

		watchesGroup := app.Group("/watches")
		watchesGroup.GET("/", watchesMine)
		watchesGroup.POST("/", watchesCreate)
		watchesGroup.DELETE("/{watch_id}", watchesRemove)

		app.POST("/upload/", uploadHandler)

		app.POST("/service", serviceHandler)

		auth := app.Group("/auth")
		auth.Middleware.Skip(setCurrentUser, authInvite, authRequest, authSelect, authCallback,
			authDestroy, serviceHandler)

		auth.POST("/invite", authInvite)

		auth.POST("/login", authRequest)
		auth.GET("/select", authSelect)

		auth.GET("/callback", authCallback)  // for Oauth
		auth.POST("/callback", authCallback) // for SAML

		auth.GET("/logout", authDestroy)

		users := app.Group("/users")
		users.GET("/me", usersMe)
		users.PUT("/me", usersMeUpdate)

		listeners.RegisterListeners()
	}

	return app
}
