dev: buffalo migrate

all: buffalo migrate adminer ppa playground

migrate: db
	docker-compose run --rm buffalo whenavail db 5432 10 buffalo-pop pop migrate up
	docker-compose run --rm buffalo grift db:seed

migratestatus: db
	docker-compose run buffalo buffalo-pop pop migrate status

gqlgen: application/gqlgen/generated.go

application/gqlgen/generated.go: application/gqlgen/schema.graphql
	docker-compose run --rm buffalo go generate ./...

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

clean:
	docker-compose kill
	docker-compose rm -f
