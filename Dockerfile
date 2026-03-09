# 第一階段：編譯
FROM golang:1.25-alpine AS builder  
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 編譯成二進位檔，關閉 CGO 以確保在 alpine 執行
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二階段：執行（極小化映像檔）
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
# 從編譯階段把檔案抓過來
COPY --from=builder /app/main .
# 暴露 Gin 預設埠
EXPOSE 8080
CMD ["./main"]