SHELL=/bin/sh
GOPATH=/go/src/github.com/Eun/docker-purge
build:
	docker run --rm --volume $(shell pwd):${GOPATH} --workdir ${GOPATH} --env GITHUB_TOKEN=${GITHUB_TOKEN} golang:1.10.3-stretch ${GOPATH}/docker-entrypoint.sh build

release:
	docker run --rm --volume $(shell pwd):${GOPATH} --workdir ${GOPATH} --env GITHUB_TOKEN=${GITHUB_TOKEN} golang:1.10.3-stretch ${GOPATH}/docker-entrypoint.sh release


interactive:
	docker run -ti --rm --volume $(shell pwd):${GOPATH} --workdir ${GOPATH} --env GITHUB_TOKEN=${GITHUB_TOKEN} golang:1.10.3-stretch bash