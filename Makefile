DEB_VERSION = $(shell dpkg-parsechangelog -S version)
GIT_SHA = $(shell git rev-parse HEAD)
GIT_TAG = $(shell git name-rev --tags --name-only ${GIT_SHA})
TIMESTAMP = $(shell date +"%s")

all: clean gansoi

gansoi:
	CGO_ENABLED=0 \
	go build \
		-ldflags " \
			-X github.com/gansoi/gansoi/build.Version=${GIT_TAG} \
			-X github.com/gansoi/gansoi/build.SHA=${GIT_SHA} \
			-X github.com/gansoi/gansoi/build.timestamp=${TIMESTAMP}" \
		-a \
		-installsuffix cgo \
		-o $@

docker-image: clean
	docker build -t gansoi/gansoi .

docker-push: docker-image
	docker push gansoi/gansoi

docker-run:
	docker run --rm -p 80:80 -p 443:443 -e DEBUG=* gansoi/gansoi

deb:
	dpkg-buildpackage -uc -B
	rm ../gansoi_$(DEB_VERSION)_amd64.changes
	mv ../gansoi_$(DEB_VERSION)_amd64.deb ./

clean:
	rm -f gansoi

test:
	go test -cover ./...

lint:
	golangci-lint run
