all: clean gansoi

gansoi:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@

dockerroot/etc/ssl/certs/ca-certificates.crt: /etc/ssl/certs/ca-certificates.crt
	cp -a $< $@

docker-image: clean gansoi dockerroot/etc/ssl/certs/ca-certificates.crt
	docker build -t abrander/gansoi .

docker-push: docker-image
	docker push abrander/gansoi

clean:
	rm -f gansoi dockerroot/etc/ssl/certs/ca-certificates.crt
