build:
	docker-compose build app

up:
	docker-compose up -d --build

down:
	docker-compose down


restart: down up

migrate-up:
	docker-compose run --rm migrate