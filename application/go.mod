module github.com/silinternational/wecarry-api

go 1.13

require (
	github.com/99designs/gqlgen v0.11.1
	github.com/aws/aws-sdk-go v1.38.13
	github.com/beevik/etree v1.1.0 // indirect
	github.com/caddyserver/certmagic v0.11.2
	github.com/fatih/color v1.10.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-acme/lego v2.7.2+incompatible
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobuffalo/buffalo v0.16.23
	github.com/gobuffalo/envy v1.9.0
	github.com/gobuffalo/events v1.4.1
	github.com/gobuffalo/fizz v1.13.0 // indirect
	github.com/gobuffalo/httptest v1.5.0
	github.com/gobuffalo/logger v1.0.3
	github.com/gobuffalo/mw-i18n v0.0.0-20191212073857-95b5d236d455
	github.com/gobuffalo/mw-paramlogger v0.0.0-20190224201358-0d45762ab655
	github.com/gobuffalo/nulls v0.4.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/gobuffalo/plush/v4 v4.1.5 // indirect
	github.com/gobuffalo/pop/v5 v5.3.4
	github.com/gobuffalo/suite/v3 v3.0.2
	github.com/gobuffalo/validate/v3 v3.3.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/pat v0.0.0-20180118222023-199c85a7f6d1
	github.com/gorilla/sessions v1.2.1
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/markbates/goth v1.56.0
	github.com/markbates/grift v1.5.0
	github.com/microcosm-cc/bluemonday v1.0.9 // indirect
	github.com/monoculum/formam v0.0.0-20210131081218-41b48e2a724b // indirect
	github.com/mrjones/oauth v0.0.0-20180629183705-f4e24b6d100c
	github.com/paganotoni/sendgrid-sender v1.0.5
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rollbar/rollbar-go v1.4.1
	github.com/rs/cors v1.6.0
	github.com/russellhaering/gosaml2 v0.3.1
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/silinternational/certmagic-storage-dynamodb v0.0.0-20200613203057-b8a4f076ab49
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.1.3 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vektah/gqlparser/v2 v2.0.1
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/image v0.0.0-20191214001246-9130b4cfad52
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	jaytaylor.com/html2text v0.0.0-20190408195923-01ec452cbe43
)

replace github.com/silinternational/wecarry-api v0.0.0 => ./

replace github.com/gobuffalo/envy v1.7.0 => github.com/silinternational/envy v1.7.0
