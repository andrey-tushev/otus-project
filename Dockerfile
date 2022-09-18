FROM alpine:3.9

COPY ./bin/previewer-linux /opt/previewer/server

RUN mkdir -p /opt/previewer/cache

WORKDIR /opt/previewer

CMD ["/opt/previewer/server", "--target-url=https://raw.githubusercontent.com/andrey-tushev/otus-go/project/project/images/www/"]

EXPOSE 8081
