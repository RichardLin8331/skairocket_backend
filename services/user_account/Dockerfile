FROM golang:1.19 AS stage1
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main 

FROM alpine:latest AS stage2
COPY ./configs/ ./configs/
COPY --from=stage1 /src/main ./
EXPOSE 8900
CMD ["./main"]