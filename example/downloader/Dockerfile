FROM alpine:latest
RUN   apk update \
  &&   apk add ca-certificates wget \
  &&   update-ca-certificates
RUN mkdir /app
ADD ./download.sh /app/
WORKDIR /app

CMD ["/bin/sh", "/app/download.sh", "/output/", "/input/*"]
