FROM golang:1.15 AS build
RUN apt-get update && apt-get install -y upx-ucl
WORKDIR /serverlessish
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w"
RUN upx serverlessish

FROM scratch
COPY --from=build /serverlessish/serverlessish /opt/extensions/serverlessish
