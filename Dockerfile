FROM golang:1.22.4-alpine

WORKDIR /course

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

WORKDIR /course/cmd/course

RUN go build -o /course/bin/course .

EXPOSE 8080
ENTRYPOINT [ "/course/bin/course" ]