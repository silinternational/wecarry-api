package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/providers/dns/cloudflare"
	"github.com/gobuffalo/buffalo/servers"
	"github.com/gobuffalo/envy"

	"github.com/rollbar/rollbar-go"

	dynamodbstore "github.com/silinternational/certmagic-storage-dynamodb"

	"github.com/silinternational/wecarry-api/actions"
	"github.com/silinternational/wecarry-api/domain"
)

var GitCommitHash string

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {

	// init rollbar
	rollbar.SetToken(domain.Env.RollbarToken)
	rollbar.SetEnvironment(domain.Env.GoEnv)
	rollbar.SetCodeVersion(GitCommitHash)
	rollbar.SetServerRoot(domain.Env.RollbarServerRoot)

	var srv servers.Server
	if domain.Env.DisableTLS {
		srv = servers.New()
	} else {
		certmagic.Default.Storage = &dynamodbstore.Storage{
			Table: "CertMagic",
		}

		cloudflareConfig := cloudflare.NewDefaultConfig()
		cloudflareConfig.AuthEmail = domain.Env.CloudflareAuthEmail
		cloudflareConfig.AuthKey = domain.Env.CloudflareAuthKey
		dnsProvider, err := cloudflare.NewDNSProviderConfig(cloudflareConfig)
		if err != nil {
			rollbar.Error("failed to init Cloudflare dns provider for LetsEncrypt: %s", err.Error())
			os.Exit(1)
		}

		if domain.Env.GoEnv != "prod" && domain.Env.GoEnv != "production" {
			certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
		}

		certmagic.DefaultACME.Email = domain.Env.SupportEmail
		certmagic.DefaultACME.Agreed = true
		certmagic.DefaultACME.DNSProvider = dnsProvider
		certmagic.DefaultACME.DisableHTTPChallenge = true
		certmagic.DefaultACME.DisableTLSALPNChallenge = true
		intPort, err := strconv.Atoi(envy.Get("PORT", "3000"))
		if err != nil {
			domain.Logger.Printf("failed to convert PORT env var to integer: %s", err.Error())
			os.Exit(2)
		}
		certmagic.HTTPSPort = intPort

		listener, err := certmagic.Listen([]string{domain.Env.CertDomainName})
		if err != nil {
			domain.Logger.Printf("failed to get TLS config: %s", err.Error())
			os.Exit(3)
		}
		srv = servers.WrapListener(&http.Server{}, listener)
	}

	app := actions.App()
	rollbar.WrapAndWait(func() {
		if err := app.Serve(srv); err != nil {
			if err.Error() != "context canceled" {
				panic(err)
			}
			os.Exit(0)
		}
	})

}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/
