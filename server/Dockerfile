FROM golang:1.20
WORKDIR /app
COPY . .

RUN go build -o server .

EXPOSE 5000
CMD ["./server"]