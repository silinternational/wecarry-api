# Welcome to the WeCarry.App server code!

[![Go Report Card](https://goreportcard.com/badge/github.com/silinternational/wecarry-api)](https://goreportcard.com/report/github.com/silinternational/wecarry-api)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/silinternational/wecarry-api/badges/quality-score.png?b=develop)](https://scrutinizer-ci.com/g/silinternational/wecarry-api/?branch=develop)
[![Codeship Status for silinternational/wecarry-api](https://app.codeship.com/projects/2ff9d1b0-8c61-0137-e598-0e4ef29cce88/status?branch=master)](https://app.codeship.com/projects/355314)

## Requirements

This software needs docker-compose version 1.24.0 at least.

## Setting it up

In an appropriate directory:
* git clone git@github.com:silinternational/wecarry-api.git
* cd wecarry-api
* cp .env.example .env

In application/grifts, copy private.go.example to private.go and 
correct the example values.

* make

Note that the data provided in `.env.example` will not allow
all features to work. In particular, the `ROLLBAR_TOKEN` must 
be valid. The fake AWS data will work for file uploads to the minIO
container, but obviously not for a real AWS S3 bucket.

## Installation Troubleshooting

Default versions which come with your operating system
may have earlier versions of these installed.

Make sure to uninstall docker-compose packages, and then following the
instructions here: https://docs.docker.com/compose/install/

## Auth

### Auth Request Error Codes
When an auth request is made to the wecarry api and something goes wrong, the api
will render json with a Code entry for the error. To see a list of possible codes, 
refer to domain/errorcodes.go.  In particular, the related codes are those 
that have a comment referring to actions.AuthRequest.

### Office365/AzureAD
To add an organization using AzureAD authentication, create a database organization record  
that includes an auth_type of `AZUREADV2` and an auth_config like the following ... 

```
{
    "TenantID": "12345678-abcd-1234-871a-940bc318789c", 
    "ClientSecret": "nice and crazy complicated secret :-)", 
    "ApplicationID": "12345678-abcd-1234-92f6-fffe9f3dfc6d"
}
```

For local development, if you are using `http`, then you will need to 
use `http:localhost` as the host for the WeCarry API, due to AzureAD's policies.
(This affects the `AUTH_CALLBACK_URL` in the `.env` file and the `buffalo.environment.HOST` value
in the docker-compose file.)

### Facebook
To add an organization using Facebook authentication, create a database organization record  
that includes an auth_type of `facebook` and an auth_config like the following ... 

```
{}
```

The two environment variables `FACEBOOK_KEY` and `FACEBOOK_SECRET` 
will need to be set for the appropriate Facebook oauth account and application.

For local development, if you are using `http`, then you will likely need to 
use `http:localhost` as the host for the WeCarry API, due to Facebook's policies.
(This affects the `AUTH_CALLBACK_URL` in the `.env` file and the `buffalo.environment.HOST` value
in the docker-compose file.)

### Google
To add an organization using Google authentication, create a database organization record  
that includes an auth_type of `GOOGLE` and an auth_config like the following ... 

```
{}
```

The two environment variables `GOOGLE_KEY` and `GOOGLE_SECRET` will need to be 
set for the appropriate Google oauth developer account. 

To learn about requirements on the Google side, start [here](https://developers.google.com/identity/protocols/OAuth2)

At this point, Google does not allow `*.local` domains to access their oauth2 api.
So, for local development, your api's host should probably just be `localhost`

(It may also be the case that using `buffalo dev` will require the use of `localhost` to avoid 
losing track of the google related session during authentication.)

### LinkedIn
To add an organization using LinkedIn authentication, create a database organization record  
that includes an auth_type of `linkedin` and an auth_config like the following ... 

```
{}
```

The two environment variables `LINKED_IN_KEY` and `LINKED_IN_SECRET` will need to be 
set for the appropriate LinkedIn oauth developer account. 

### SAML
To enable authentication via a SAML2 Identity Provider, an organization 
record will need to be created that includes an auth_type of `SAML` and an
auth_config like the following ...

```
{
 "SPEntityID": "http://example.local:3000", 
 "AudienceURI": "http://example.local:3000", 
 "IDPEntityID": "our.idp.net", 
 "SignRequest": true, 
 "AttributeMap": null, 
 "SPPrivateKey": "-----BEGIN PRIVATE KEY-----\nMIIG/gIB...OJxmEMBgT\n-----END PRIVATE KEY-----\n", 
 "SPPublicCert": "-----BEGIN CERTIFICATE-----\nMIIEXTCC...xmvKt42A=\n-----END CERTIFICATE-----\n", 
 "IDPPublicCert": "-----BEGIN CERTIFICATE-----\nMIIDXTC...2bb\nPw==\n-----END CERTIFICATE-----", 
 "SingleLogoutURL": "https://our.idp.net/saml2/idp/SingleLogoutService.php", 
 "SingleSignOnURL": "https://our.idp.net/saml2/idp/SSOService.php", 
 "CheckResponseSigning": true, 
 "RequireEncryptedAssertion": false, 
 "AssertionConsumerServiceURL": "http://example.local:3000/auth/callback"
}
```

### Twitter (Dicey Auth Option)
To add an organization using Twitter authentication, create a database organization record  
that includes an auth_type of `twitter` and an auth_config like the following ... 

```
{}
```

The two environment variables `TWITTER_KEY` and `TWITTER_SECRET` will need to be 
set for the appropriate LinkedIn oauth developer account. 

The problem with Twitter is that its users don't necessarily have a separate 
First Name and Last Name. We added a function that either uses a space or 
underscore as the separator (based on the User.Name) or just duplicates the
User.Name as both the First and Last Names.

## GraphQL API

### API Documentation

Reference the GraphQL Schema at application/gqlgen/schema.graphql or use
GraphQL introspection to access the schema from the running app. API tools such
as [Insomnia](https://insomnia.rest) include an interactive schema browser that 
make use of GraphQL introspection.
 
### Conventions

#### Implicit `null` vs explicit `null

The standard established for this API is both implicit null (field
omitted from mutation) and explicit null (`null` specified in mutation) 
will erase or set to `null`. Note that this does not address the
GraphQL spec requirement to not modify an omitted field. The reference UI 
implementation, [wecarry-ui](https://github.com/silinternational/wecarry-ui),
will always include all supported fields.
