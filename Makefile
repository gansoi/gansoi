DEB_VERSION = $(shell dpkg-parsechangelog -S version)

all: clean gansoi

gansoi:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@

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
