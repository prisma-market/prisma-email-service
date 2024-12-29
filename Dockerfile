# 빌드 스테이지
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 의존성 파일 복사 및 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스코드 복사
COPY . .

# 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# 실행 스테이지
FROM alpine:latest

WORKDIR /app

# 타임존 설정을 위한 패키지 설치
RUN apk add --no-cache tzdata

# 빌드된 바이너리 복사
COPY --from=builder /app/main .
COPY .env .

# 서비스 포트 노출
EXPOSE 8003

CMD ["./main"]