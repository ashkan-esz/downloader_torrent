update_swagger:
	swag init -g cmd/api/main.go --output docs

run_redis:
	docker run --restart unless-stopped -d --network=host --memory 200m -e ALLOW_EMPTY_PASSWORD=yes redis:alpine

run_dev:
	clear && swag fmt && make update_swagger && go run cmd/api/main.go

up:
	docker-compose up --build

build:
	docker image build --network=host -t downloader_torrent .

run:
	docker run --rm --network=host --name downloader_torrent -v ./downloads:/app/downloads --env-file ./.env downloader_torrent

push-image:
	docker tag downloader_torrent ashkanaz2828/downloader_torrent
	docker push ashkanaz2828/downloader_torrent

.PHONY: db_dev create_db_dev drop_db_dev migrate_up_db_dev migrate_down_db_dev update_swagger run_redis run_dev up build run push-image