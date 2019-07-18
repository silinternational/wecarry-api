dev: buffalo migrate

migrate: db
	docker-compose run --rm buffalo whenavail db 5432 10 buffalo-pop pop migrate up
	docker-compose run --rm buffalo grift db:seed

migratestatus: db
	docker-compose run buffalo buffalo-pop pop migrate status

gqlgen: application/gqlgen/generated.go

application/gqlgen/generated.go: application/gqlgen/schema.graphql
	docker-compose run --rm buffalo go generate ./...

buffalo: db
	docker-compose up -d buffalo

db:
	docker-compose up -d db

clean:
	docker-compose kill
	docker-compose rm -f
