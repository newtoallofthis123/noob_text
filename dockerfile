FROM golang:1.21-alpine as builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download
RUN go install github.com/a-h/templ

COPY *.go .
COPY ./templates ./templates
COPY ./public ./public
COPY ./.env ./.env

RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o ./noobtext

FROM scratch

COPY --from=builder /app/noobtext /app/noobtext
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/public /app/public
COPY --from=builder /app/.env /app/.env
WORKDIR /app
CMD [ "./noobtext" ]
EXPOSE 3579
