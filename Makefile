colima_docker_sock = unix:///$$HOME/.config/colima/default/docker.sock

.PHONY: start-dev
start-dev:
	DOCKER_HOST=$(colima_docker_sock) go run . $(flags)

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: update-deps
update-deps: mod-tidy
	go get -u ./...

.PHONY: clean
clean:
	rm -rf $$XDG_STATE_HOME/minit/minit-package-store/
	docker system prune --all --force
