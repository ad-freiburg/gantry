FROM alpine:latest
RUN mkdir /app
ADD ./cat.sh /app/

CMD ["/bin/sh", "/app/cat.sh", "/input/*"]
