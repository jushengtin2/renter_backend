# 租屋平台後端（Renter Backend）
網址：https://renter-nine.vercel.app/

以 Go + Gin + GORM 開發的租屋平台後端服務，提供貼文、留言、按讚/收藏、使用者同步（Clerk webhook）等功能，資料庫使用 PostgreSQL（可連本地 PostgreSQL 或 Supabase Postgres）。

## 專案定位

- 提供租屋貼文與留言相關 REST API
- 串接 Clerk JWT 做身分驗證
- 串接 Google Cloud Storage（GCS）儲存貼文/留言圖片
- 可部署到 GCP Cloud Run（已附 GitHub Actions workflow）

## 程式架構（3 層）

本專案採用 `Controller -> Service -> Model` 分層：

1. Controller 層（`internal/controllers`）

- 負責 HTTP 請求解析、參數驗證、呼叫 Service、回傳 JSON
- 不直接寫商業邏輯與 SQL

2. Service 層（`internal/services`）

- 核心商業邏輯
- 包含查詢排序、篩選、資料組裝、按讚/收藏計數更新、圖片上傳 GCS

3. Model 層（`internal/models`）

- GORM model 定義資料表結構
- 對應 PostgreSQL 欄位型別（含 `jsonb`、`geography(Point,4326)`、軟刪除）

4. 基礎設施層（`internal/database`, `internal/middleware`）

- 資料庫連線、AutoMigrate、JWT 驗證中介層

## 專案目錄

```text
.
├── main.go
├── internal
│   ├── router.go
│   ├── controllers
│   ├── services
│   ├── models
│   ├── database
│   └── middleware
├── .github/workflows/backend.yml
├── Dockerfile
└── terraform/
```

## 核心功能

### 1) 貼文系統

- 建立貼文（含多張圖片）
- 取得首頁貼文列表
- 取得單一貼文
- 刪除貼文（軟刪除）
- 貼文按讚 / 取消按讚
- 貼文收藏 / 取消收藏
- 取得我的收藏貼文

查詢能力：

- 排序：`hot` / `new` / `old`
- 文字搜尋：標題/內容模糊查詢 + `word_similarity`
- 地區篩選：台北市 / 新北市行政區
- 地圖框選：利用 PostGIS `ST_MakeEnvelope` + `geography` 欄位
- 分頁：`page` + `page_size`（offset pagination）

### 2) 留言系統

- 新增留言（可回覆留言形成樹狀結構）
- 取得貼文留言（支援 `popular` / `new` / `old`）
- 留言按讚 / 取消按讚
- 刪除留言（軟刪除）

### 3) 使用者同步（Clerk webhook）

- 支援 `user.created` / `user.updated` / `user.deleted`
- 將 Clerk 使用者資料同步到本地 `users` 表

### 4) 圖片儲存

- 貼文與留言圖片上傳到 GCS
- DB 以 `jsonb` 儲存圖片 URL 陣列

### 5) 時間處理

- 建立貼文與留言時，以 `UTC` 寫入 `CreatedAt`

## 技術棧

- 語言：Go 1.25
- Web Framework：Gin
- ORM：GORM
- DB：PostgreSQL / Supabase Postgres
- Auth：Clerk（JWKS 驗證 JWT）
- Storage：Google Cloud Storage
- Deployment：Docker + Cloud Run + GitHub Actions
