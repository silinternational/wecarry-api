package actions

// WeCarry API
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//  Schemes: https
//  Host: localhost
//  BasePath: /
//  Version: 0.0.1
//  License: MIT http://opensource.org/licenses/MIT
//
//  Consumes:
//  - application/json
//
//  Produces:
//  - application/json
//
//  Security:
//  - oauth2:
//
//  SecurityDefinitions:
//  bearerAuth:
//      type: http
//      scheme: bearer
//
// swagger:meta

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	i18n "github.com/gobuffalo/mw-i18n/v2"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/rs/cors"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
	"github.com/silinternational/wecarry-api/listeners"
	"github.com/silinternational/wecarry-api/locales"
	"github.com/silinternational/wecarry-api/log"
	"github.com/silinternational/wecarry-api/models"
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
			Env:         domain.Env.GoEnv,
			SessionName: "_wecarry_session",
			PreWares: []buffalo.PreWare{
				cors.New(cors.Options{
					AllowCredentials: true,
					AllowedOrigins:   []string{domain.Env.UIURL},
					AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
					AllowedHeaders:   []string{"*"},
				}).Handler,
			},
		})

		// Setup and use translations. This should be the first middleware if all other middleware use i18n
		app.Use(translations())

		registerCustomErrorHandler(app)

		// Initialize remote logger middleware
		app.Use(log.SentryMiddleware)

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		//  Added for authorization
		app.Use(setCurrentUser)
		app.Middleware.Skip(setCurrentUser, statusHandler, serviceHandler)

		// Wraps each request in a transaction.
		app.Use(popmw.Transaction(models.DB))

		app.GET("/site/status", statusHandler)
		app.Middleware.Skip(buffalo.RequestLogger, statusHandler)

		eventsGroup := app.Group("/events")
		eventsGroup.GET("/", meetingsList)
		eventsGroup.POST("/", meetingsCreate)
		eventsGroup.PUT("/{event_id}", meetingsUpdate)
		eventsGroup.POST("/join", meetingsJoin)
		eventsGroup.GET("/{event_id}", meetingsGet)
		eventsGroup.DELETE("/{event_id}", meetingsRemove)
		eventsGroup.DELETE("/{event_id}/invite/", meetingsInviteDelete)

		app.POST("/messages/", messagesCreate)

		threadsGroup := app.Group("/threads")
		threadsGroup.GET("/", threadsMine)
		threadsGroup.PUT("/{thread_id}/read", threadsMarkAsRead)

		requestsGroup := app.Group("/requests")
		requestsGroup.GET("/", requestsList)
		requestsGroup.POST("/", requestsCreate)
		requestsGroup.GET("/{request_id}", requestsGet)
		requestsGroup.PUT("/{request_id}", requestsUpdate)
		requestsGroup.PUT("/{request_id}/status", requestsUpdateStatus)

		requestsGroup.POST("/{request_id}/potentialprovider", requestsAddMeAsPotentialProvider)
		requestsGroup.DELETE("/{request_id}/potentialprovider/{user_id}", requestsRejectPotentialProvider)
		requestsGroup.DELETE("/{request_id}/potentialprovider", requestsRemoveMeAsPotentialProvider)

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

		listeners.RegisterListener()

		job.Init(&app.Worker)
	}

	return app
}

// translations will load locale files, set up the translator `domain.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if domain.T, err = i18n.New(locales.FS(), "en"); err != nil {
		_ = app.Stop(err)
	}
	return domain.T.Middleware()
}
