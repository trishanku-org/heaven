FROM golang:1.19-alpine3.17 AS builder

WORKDIR /trishanku/heaven

COPY api api/
COPY config config/
COPY controllers controllers/
COPY pkg pkg/
COPY vendor vendor/
COPY go.mod go.sum main.go VERSION ./

RUN go build -mod=vendor \
		-o /heaven \
		-ldflags "-X github.com/trishanku/heaven/controllers.Version=$(cat VERSION)" \
		main.go

FROM alpine:3.17 AS runner

COPY --from=builder /heaven /heaven

COPY --from=builder /trishanku/heaven/config /etc/trishanku/heaven/config

ENTRYPOINT [ "/heaven" ]
