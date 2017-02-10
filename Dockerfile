FROM scratch

COPY dockerroot /
COPY gansoi /
COPY web /web

ENV DEBUG=*
ENV GANSOI_WEBROOT=/web

VOLUME /data

ENTRYPOINT [ "/gansoi" ]
CMD [ "core", "init-and-run" ]
