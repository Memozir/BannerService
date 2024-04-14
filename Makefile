# Makefile

docker_all:
	docker-compose -f ./deployments/docker-compose.yml up -d

migrations_up:
	cd ./migrations && goose up
