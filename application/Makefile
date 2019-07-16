dev: buffalo migrate

migrate: db
	docker-compose run buffalo buffalo-pop pop migrate up

migratestatus: db
	docker-compose run buffalo buffalo-pop pop migrate status

buffalo: db
	docker-compose up -d buffalo

db:
	docker-compose up -d db

clean:
	docker-compose kill
	docker-compose rm -f