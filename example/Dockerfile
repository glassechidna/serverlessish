FROM golang:1.15 AS build
WORKDIR /example
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w"

FROM public.ecr.aws/c2t6n2x5/serverlessish:1 AS s

FROM gcr.io/distroless/static
COPY --from=s /opt/extensions/serverlessish /opt/extensions/serverlessish
COPY --from=build /example/example /example
ENTRYPOINT ["/example"]

#FROM public.ecr.aws/c2t6n2x5/serverlessish:1 AS s
#
#FROM golang:1.15-alpine
#COPY --from=s /opt/extensions/serverlessish /opt/extensions/serverlessish
#
#WORKDIR /example
#COPY . .
#RUN go build -ldflags="-s -w"
#ENTRYPOINT ["/example/example"]
