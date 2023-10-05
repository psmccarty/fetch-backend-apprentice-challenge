FROM golang:1.20
WORKDIR /app
COPY . /app
RUN apt-get update && go mod tidy && go mod vendor
EXPOSE 8080
CMD ["go", "run", "cmd/application/main.go"]