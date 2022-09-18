# Примеры команд

## Запуск с дефолтными параметрами
go run ./cmd
Целевой путь будет http://localhost:8082/

## Проксирование картинок из github
go run ./cmd --target-url=https://raw.githubusercontent.com/andrey-tushev/otus-go/previewer/previewer/images/

## Запуск локального хостинга картинок на nginx
docker run --name img-srv --rm -v $(shell pwd)/images/www:/usr/share/nginx/html -p 8082:80 -d nginx:alpine

## Makefile

Также смотри Makefile c таргетами под разные запуски, сборки и тесты

# Хостинг картинок

* `http://127.0.0.1:8082/<img>.jpg` (локальный docker run)
* `http://127.0.0.1:8083/<img>.jpg` (локальный docker-compose)
* `https://raw.githubusercontent.com/andrey-tushev/otus-go/project/project/images/www/<img>.jpg` (внешний хостинг)

где `<img>` может быть:
* cat-1.jpg
* cat-2.jpg
* cat-3.jpg
* cat-4.jpg
* cat-5.jpg
* bad.jpg - битая картинка

# Ручная проверка прокси-превьювера

`http://127.0.0.1:8081/fill/300/200/<img>.jpg`
