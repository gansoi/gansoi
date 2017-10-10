FROM golang:1.9
WORKDIR /go/src/github.com/gansoi/gansoi
COPY . /go/src/github.com/gansoi/gansoi

RUN go get ./...
RUN make gansoi

FROM scratch

COPY dockerroot /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/gansoi/gansoi/gansoi /
COPY web /web

ENV DEBUG=*
ENV GANSOI_WEBROOT=/web

VOLUME /data

ENTRYPOINT [ "/gansoi" ]
CMD [ "core", "init-and-run" ]
