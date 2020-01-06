# Welcome to the WeCarry.App server code!

## Requirements

This software needs docker-compose version 1.24.0 at least.

## Setting it up

In an appropriate directory:
* git clone git@github.com:silinternational/wecarry-api.git
* cd wecarry-api
* cp .env.example .env

In application/grifts, copy private.example.go to private.go and 
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

### Google
To enable authentication via Google, an organization record will 
need to be created that includes an auth_type of `google` and an auth_config like the following ... 

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

### SAML
To enable authentication via a SAML2 Identity Provider, an organization 
record will need to be created that includes an auth_type of `saml` and an
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
