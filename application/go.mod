module github.com/silinternational/wecarry-api

go 1.13

require (
	github.com/99designs/gqlgen v0.10.1
	github.com/aws/aws-sdk-go v1.25.0
	github.com/beevik/etree v1.1.0 // indirect
	github.com/caddyserver/certmagic v0.10.5
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cockroachdb/cockroach-go v0.0.0-20190925194419-606b3d062051 // indirect
	github.com/go-acme/lego v2.7.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/gobuffalo/buffalo v0.15.3
	github.com/gobuffalo/envy v1.8.1
	github.com/gobuffalo/events v1.4.0
	github.com/gobuffalo/httptest v1.4.0
	github.com/gobuffalo/logger v1.0.3
	github.com/gobuffalo/mw-i18n v0.0.0-20191212073857-95b5d236d455
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/nulls v0.1.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gobuffalo/pop v4.12.2+incompatible
	github.com/gobuffalo/suite v2.8.1+incompatible
	github.com/gobuffalo/validate v2.0.4+incompatible
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/pat v0.0.0-20180118222023-199c85a7f6d1
	github.com/gorilla/sessions v1.2.0
	github.com/jackc/pgconn v1.1.0 // indirect
	github.com/jackc/pgx v3.5.0+incompatible // indirect
	github.com/markbates/goth v1.56.0
	github.com/markbates/grift v1.5.0
	github.com/mrjones/oauth v0.0.0-20180629183705-f4e24b6d100c
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/paganotoni/sendgrid-sender v1.0.5
	github.com/pkg/errors v0.8.1
	github.com/rollbar/rollbar-go v1.1.0
	github.com/rs/cors v1.6.0
	github.com/russellhaering/gosaml2 v0.3.1
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/silinternational/certmagic-storage-dynamodb v0.0.0-20200316155358-4e6c577f652a
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.1.2
	golang.org/x/image v0.0.0-20191214001246-9130b4cfad52
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	jaytaylor.com/html2text v0.0.0-20190408195923-01ec452cbe43
)

replace github.com/silinternational/wecarry-api v0.0.0 => ./

replace github.com/gobuffalo/envy v1.7.0 => github.com/silinternational/envy v1.7.0
