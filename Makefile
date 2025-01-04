colima_docker_sock = unix:///$$HOME/.config/colima/default/docker.sock

.PHONY: start-dev
start-dev:
	DOCKER_HOST=$(colima_docker_sock) go run .

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: update-deps
update-deps: mod-tidy
	go get -u ./...
