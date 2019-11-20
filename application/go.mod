module github.com/silinternational/wecarry-api

go 1.12

require (
	github.com/99designs/gqlgen v0.10.1
	github.com/aws/aws-sdk-go v1.25.0
	github.com/beevik/etree v1.1.0 // indirect
	github.com/go-chi/chi v3.3.2+incompatible
	github.com/gobuffalo/buffalo v0.14.6
	github.com/gobuffalo/buffalo-pop v1.16.0
	github.com/gobuffalo/envy v1.7.0
	github.com/gobuffalo/events v1.4.0
	github.com/gobuffalo/httptest v1.4.0
	github.com/gobuffalo/mw-csrf v0.0.0-20190129204204-25460a055517
	github.com/gobuffalo/mw-i18n v0.0.0-20190129204410-552713a3ebb4
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/nulls v0.1.0
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gobuffalo/pop v4.11.2+incompatible
	github.com/gobuffalo/suite v2.8.1+incompatible
	github.com/gobuffalo/validate v2.0.3+incompatible
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/sessions v1.1.3
	github.com/markbates/goth v1.56.0
	github.com/markbates/grift v1.0.6
	github.com/nicksnyder/go-i18n v1.10.0
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/paganotoni/sendgrid-sender v1.0.5
	github.com/pkg/errors v0.8.1
	github.com/rollbar/rollbar-go v1.1.0
	github.com/rs/cors v1.6.0
	github.com/russellhaering/gosaml2 v0.3.1
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.1.2
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890
	jaytaylor.com/html2text v0.0.0-20190408195923-01ec452cbe43
)

replace github.com/silinternational/wecarry-api v0.0.0 => ./

replace github.com/gobuffalo/envy v1.7.0 => github.com/silinternational/envy v1.7.0
