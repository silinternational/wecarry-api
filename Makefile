dev: buffalo migrate adminer

all: buffalo migrate adminer ppa playground

migrate: db
	docker-compose run --rm buffalo whenavail db 5432 10 buffalo-pop pop migrate up
	docker-compose run --rm buffalo /bin/bash -c "grift private:seed && grift db:seed && grift minio:seed"

migratestatus: db
	docker-compose run buffalo buffalo-pop pop migrate status

migratetestdb: testdb
	docker-compose run --rm test whenavail testdb 5432 10 buffalo-pop pop migrate up

gqlgen: application/gqlgen/generated.go

application/gqlgen/generated.go: application/gqlgen/schema.graphql application/gqlgen/gqlgen.yml
	docker-compose run --rm buffalo /bin/bash -c "go generate ./gqlgen ; chown 1000.1000 gqlgen/generated.go gqlgen/models_gen.go"

adminer:
	docker-compose up -d adminer

playground:
	docker-compose up -d playground

ppa:
	docker-compose up -d phppgadmin

buffalo: db
	docker-compose up -d buffalo

bounce: db
	docker-compose kill buffalo
	docker-compose rm buffalo
	docker-compose up -d buffalo

logs:
	docker-compose logs buffalo

db:
	docker-compose up -d db

testdb:
	docker-compose up -d testdb

test:
	docker-compose run --rm test whenavail testdb 5432 10 buffalo test

testenv: migratetestdb
	docker-compose run --rm test bash

clean:
	docker-compose kill
	docker-compose rm -f

fresh: clean dev
