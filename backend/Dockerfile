# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# 의존성 먼저 복사
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/dashboard-server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 리소스 파일들도 복사
COPY --from=builder /app/internal/resources ./internal/resources

ENV TZ=Asia/Seoul

EXPOSE 8080

CMD ["./main"]