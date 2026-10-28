FROM golang:1.21.4-alpine AS build

WORKDIR /shux-api
COPY . /shux-api

RUN go mod tidy
RUN go build -o /bin/shux-api ./main.go

FROM alpine:latest

COPY --from=build /bin/shux-api /bin/shux-api

ENV API_PORT="3000"
EXPOSE 3000

CMD ["/bin/shux-api"]
