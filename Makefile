.PHONY:
docker-postgres:
	docker run -p 5432:5432 -e POSTGRES_USER=investment_game_backend -e POSTGRES_PASSWORD=investment_game_backend -e POSTGRES_DB=investment_game_backend -d --name pg-local postgres

.PHONY:
docker-app:
	docker compose up -d --build