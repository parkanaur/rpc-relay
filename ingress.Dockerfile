FROM golang:1.18

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN ["sh", "scripts/buildsrv", "ingress"]

CMD ["sh", "scripts/startsrv", "ingress"]