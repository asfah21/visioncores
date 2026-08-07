FROM golang:1.22

WORKDIR /app
COPY . .

RUN go mod init backend || true
RUN go get github.com/gofiber/fiber/v2
RUN go get github.com/gofiber/websocket/v2
RUN go get github.com/lib/pq

RUN go build -o main .

CMD ["./main"]

