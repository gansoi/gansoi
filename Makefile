DEB_VERSION = $(shell dpkg-parsechangelog -S version)
GIT_SHA = $(shell git rev-parse HEAD)
GIT_TAG = $(shell git name-rev --tags --name-only ${GIT_SHA})
TIMESTAMP = $(shell date +"%s")

all: clean gansoi

gansoi:
	CGO_ENABLED=0 GOOS=linux \
	go build \
		-ldflags " \
			-X github.com/gansoi/gansoi/build.Version=${GIT_TAG} \
			-X github.com/gansoi/gansoi/build.SHA=${GIT_SHA} \
			-X github.com/gansoi/gansoi/build.timestamp=${TIMESTAMP}" \
		-a \
		-installsuffix cgo \
		-o $@

dockerroot/etc/ssl/certs/ca-certificates.crt: /etc/ssl/certs/ca-certificates.crt
	cp -a $< $@

docker-image: clean gansoi dockerroot/etc/ssl/certs/ca-certificates.crt
	docker build -t abrander/gansoi .

docker-push: docker-image
	docker push abrander/gansoi

docker-run:
	docker run --rm -p 80:80 -p 443:443 -e DEBUG=* abrander/gansoi

deb:
	dpkg-buildpackage -uc -B
	rm ../gansoi_$(DEB_VERSION)_amd64.changes
	mv ../gansoi_$(DEB_VERSION)_amd64.deb ./

clean:
	rm -f gansoi dockerroot/etc/ssl/certs/ca-certificates.crt

test:
	go test -cover ./...
