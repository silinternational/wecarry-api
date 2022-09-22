dev: buffalo adminer migrate 

migrate:
	docker-compose run --rm buffalo whenavail db 5432 10 buffalo-pop pop migrate up
	docker-compose run --rm buffalo /bin/bash -c "buffalo task private:seed && buffalo task db:seed && buffalo task minio:seed"

migratestatus:
	docker-compose run buffalo buffalo-pop pop migrate status

migratetestdb: testdb
	docker-compose run --rm test whenavail testdb 5432 10 buffalo-pop pop migrate up

adminer:
	docker-compose up -d adminer

buffalo:
	docker-compose up -d buffalo

redis: db
	docker-compose up -d redis
	
debug: killbuffalo
	docker-compose up -d debug
	docker-compose logs -f debug

swagger: swaggerspec
	docker-compose run --rm --service-ports swagger serve -p 8082 --no-open swagger.json

swaggerspec:
	docker-compose run --rm swagger generate spec -m -o swagger.json

bounce:
	docker-compose kill buffalo
	docker-compose rm -f buffalo
	docker-compose up -d buffalo

logs:
	docker-compose logs buffalo

testdb:
	docker-compose up -d testdb

test:
	docker-compose run --rm test whenavail testdb 5432 10 buffalo test

testenv: rmtestdb migratetestdb
	@echo "\n\nIf minio hasn't been initialized, run buffalo task minio:seed\n"
	docker-compose run --rm test bash

rmtestdb:
	docker-compose kill testdb && docker-compose rm -f testdb

killbuffalo:
	docker-compose kill buffalo

clean:
	docker-compose kill
	docker-compose rm -f

fresh: clean dev
