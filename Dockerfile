#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY . .
RUN go get -d -v ./...
RUN go build -o app

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/app /app
ENV MONGODB_URI=""
ENV JWT_VALIDATION_URI=""
ENV ADMIN_HYDRA_HOST=""
ENV IPFS_API_SERVER_URI=""
ENV IPFS_GATEWAY_URI=""
ENTRYPOINT ./app
LABEL Name=ipfsapi Version=0.0.1
EXPOSE 4000
