package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"renter_backend/internal/models"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 只是定義 PostService 是包了一個db的結構 負責文章的商業邏輯
type PostService struct {
	db        *gorm.DB
	//rdb       *redis.Client
	gcsClient *storage.Client
}

var allowedDistrictsByCity = map[string]map[string]struct{}{
	"台北市": buildDistrictSet(
		"中正區", "大同區", "中山區", "松山區", "大安區", "萬華區",
		"信義區", "士林區", "北投區", "內湖區", "南港區", "文山區",
	),
	"新北市": buildDistrictSet(
		"板橋區", "三重區", "中和區", "永和區", "新莊區", "新店區",
		"土城區", "蘆洲區", "樹林區", "鶯歌區", "三峽區", "淡水區",
		"汐止區", "瑞芳區", "五股區", "泰山區", "林口區", "深坑區",
		"石碇區", "坪林區", "三芝區", "石門區", "八里區", "平溪區",
		"雙溪區", "貢寮區", "金山區", "萬里區", "烏來區",
	),
}

func buildDistrictSet(districts ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(districts))
	for _, district := range districts {
		set[district] = struct{}{}
	}
	return set
}

// NewPostService 用來建立 PostService 接收 DB 連線
func NewPostService(db *gorm.DB, gcsClient *storage.Client) *PostService {
	return &PostService{db: db,  gcsClient: gcsClient}
}

type MainPagePostResponse struct {
	PostID         int      `json:"post_id"`
	Title          string   `json:"title"`
	Timestamp      string   `json:"timestamp"`
	PictureURL     []string `json:"picture_url"`
	Content        string   `json:"content"`
	LikeNumber     int      `json:"like_number"`
	SaveNumber     int      `json:"save_number"`
	BookmarkNumber int      `json:"bookmark_number"`
	Tags           []string `json:"tags"`
	ILikeThis      bool     `json:"i_like_this"`
	ISaveThis      bool     `json:"i_save_this"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	City      string         `json:"city"`
	District  string         `json:"district"`
}

type MainPagePostPageResponse struct {
	Posts       []*MainPagePostResponse `json:"posts"`
	TotalCount  int                     `json:"total_count"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	HasNextPage bool                    `json:"has_next_page"`
}

type SavePostPageResponse struct {
	PostID         int      `json:"post_id"`
	Title          string   `json:"title"`
	Timestamp      string   `json:"timestamp"`
	PictureURL     []string `json:"picture_url"`
	Content        string   `json:"content"`
	City			string 	`json:"city"`
	District		string	`json:"district"`
}

// GetPostByID 取單一文章
// 不只要去post table抓 還要去post_tags, tags table抓
type PostResponse struct {
	PostID         int            `json:"post_id"`
	PictureURL     []string       `json:"picture_url"`
	Author         string         `json:"author"`
	AuthorID       string         `json:"author_id"`
	AuthorPic      string         `json:"author_pic"`
	Title          string         `json:"title"`
	Timestamp      string         `json:"timestamp"`
	Content        string         `json:"content"`
	LikeNumber     int            `json:"like_number"`
	BookmarkNumber int            `json:"bookmark_number"`
	Tags           datatypes.JSON `json:"tags"`
	ILikeThis      bool           `json:"i_like_this"`
	ISaveThis      bool           `json:"i_save_this"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	Address 		string			`json:"address"`
}

// GetPostForMainPage 取熱門文章
func (ps *PostService) GetPostForMainPage(
	sort string,
	searchText string,
	selected_area string,
	selected_tag []string,
	userID string,
	page int,
	pageSize int,
	minLat float64,
	maxLat float64,
	minLng float64,
	maxLng float64,
) (*MainPagePostPageResponse, error) {
	var posts []models.Post
	query := ps.db.Model(&models.Post{})

	// 排序方式
	var orderBy string
	switch sort {
	case "hot":
		// 讚數相同時，ID 較大(較新)的排前面
		orderBy = "like_number DESC, post_id DESC"
	case "old":
		orderBy = "post_id ASC"
	case "new":
		orderBy = "post_id DESC"
	default:
		orderBy = "like_number DESC, post_id DESC"
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize //因為我學airbnb所以不用cursor-based pagination 直接用offset就好 這樣前端要跳到第幾頁就直接算出offset傳給我就好

	// 1. 處理地區篩選
	area := strings.TrimSpace(selected_area)
	if area != "" && area != "all" {
		query = query.Where("district = ?", area)
	}

	// 1.1 文字搜尋（使用 posts.search_vector）
	trimmedSearchText := strings.TrimSpace(searchText)
	if trimmedSearchText != "" {
		likePattern := "%" + trimmedSearchText + "%"
		query = query.Where(
			"(search_vector @@ plainto_tsquery('chinese_zh', ?) OR title ILIKE ?)",
			trimmedSearchText,
			likePattern,
		)

		// 全文命中優先，其餘再依原本 sort 規則
		query = query.Order(clause.Expr{
			SQL:  "search_vector @@ plainto_tsquery('chinese_zh', ?) DESC",
			Vars: []interface{}{trimmedSearchText},
		})
	}

	// 2. 處理地圖經緯度範圍篩選（PostGIS point: location）
	if !math.IsNaN(minLat) && !math.IsNaN(maxLat) && !math.IsNaN(minLng) && !math.IsNaN(maxLng) {
		query = query.Where(
			"location IS NOT NULL AND location::geometry && ST_MakeEnvelope(?, ?, ?, ?, 4326)",
			minLng, minLat, maxLng, maxLat,
		)
	}

	// 3. 處理標籤篩選
	trimmedTags := make([]string, 0, len(selected_tag))
	for _, tag := range selected_tag {
		if t := strings.TrimSpace(tag); t != "" {
			trimmedTags = append(trimmedTags, t)
		}
	}

	if len(trimmedTags) > 0 {
		conditions := make([]string, 0, len(trimmedTags))
		args := make([]interface{}, 0, len(trimmedTags))
		for _, tag := range trimmedTags {
			tagJSON, err := json.Marshal([]string{tag})
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, "tags @> ?::jsonb")
			args = append(args, string(tagJSON))
		}
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	// 4. 先算總筆數
	var totalCount int64
	if err := query.Session(&gorm.Session{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 5. offset 分頁查詢
	if err := query.Order(orderBy).Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return &MainPagePostPageResponse{
			Posts:       []*MainPagePostResponse{},
			TotalCount:  int(totalCount),
			Page:        page,
			PageSize:    pageSize,
			HasNextPage: false,
		}, nil
	}

	// 整理 postIDs
	postIDs := make([]int, len(posts))
	for i, p := range posts {
		postIDs[i] = p.PostID
	}

	// -------------------------
	//  查詢使用者是否按讚
	// -------------------------
	likedMap := make(map[int]bool)
	if userID != "" {
		var likedIDs []int
		ps.db.Table("post_likes").
			Select("post_id").
			Where("user_id = ? AND post_id IN ?", userID, postIDs).
			Pluck("post_id", &likedIDs)

		for _, id := range likedIDs {
			likedMap[id] = true
		}
	}

	// -------------------------
	//  查詢是否收藏
	// -------------------------
	savedMap := make(map[int]bool)
	if userID != "" {
		var saveIDs []int
		ps.db.Table("post_saves").
			Select("post_id").
			Where("user_id = ? AND post_id IN ?", userID, postIDs).
			Pluck("post_id", &saveIDs)

		for _, id := range saveIDs {
			savedMap[id] = true
		}
	}

	// -------------------------
	// 組合 response DTO
	// -------------------------
	res := make([]*MainPagePostResponse, 0, len(posts))

	for _, p := range posts {

		// 轉換 JSONB pictureURL → []string（若需要）
		var pics []string
		if len(p.Picture) > 0 {
			json.Unmarshal(p.Picture, &pics)
		}

		res = append(res, &MainPagePostResponse{
			PostID:         p.PostID,
			Title:          p.Title,
			Timestamp:      p.CreatedAt.Format("2006-01-02 15:04:05"),
			PictureURL:     pics,
			Content:        p.Content,
			LikeNumber:     p.LikeNumber,
			SaveNumber:     p.SaveNumber,
			BookmarkNumber: p.SaveNumber,
			Tags:           nil,                // 目前不用 tags
			ILikeThis:      likedMap[p.PostID], // O(1)
			ISaveThis:      savedMap[p.PostID], // O(1)
			Latitude:       p.Latitude,
			Longitude:      p.Longitude,
			City:            p.City,
			District:      	p.District,
			
		})
	}

	return &MainPagePostPageResponse{
		Posts:       res,
		TotalCount:  int(totalCount),
		Page:        page,
		PageSize:    pageSize,
		HasNextPage: offset+len(posts) < int(totalCount),
	}, nil
}

func (ps *PostService) GetPostByID(postID string, userID string) (*PostResponse, error) {
	var post models.Post
	var author models.User

	var ilikethis bool
	var isavethis bool

	// 撈文章
	if err := ps.db.First(&post, "post_id = ?", postID).Error; err != nil {
		return nil, err
	}
	if err := ps.db.First(&author, "user_id = ?", post.UserID).Error; err != nil {
		return nil, err
	}

	var pics []string //因為在postgre存的是jsonb格式 但go沒這種格式 所以要轉成string array
	if len(post.Picture) > 0 {
		if err := json.Unmarshal(post.Picture, &pics); err != nil {
			return nil, err
		}
	}

	//

	// user 已登入
	if userID != "" {
		// 是否按讚
		{
			var exists bool
			err := ps.db.Model(&models.PostLike{}).
				Select("1").
				Where("user_id = ? AND post_id = ?", userID, postID).
				Limit(1).
				Scan(&exists).Error

			if err != nil {
				return nil, err
			}

			ilikethis = exists
			fmt.Println("yes or no:", ilikethis)
		}

		// 是否收藏
		{
			var exists bool
			err := ps.db.Model(&models.PostSave{}).
				Select("1").
				Where("user_id = ? AND post_id = ?", userID, postID).
				Limit(1).
				Scan(&exists).Error

			if err != nil {
				return nil, err
			}

			isavethis = exists
		}

	} else {
		ilikethis = false
		isavethis = false
	}

	//這邊會去看使用者有沒有對這篇文章按讚或收藏 然後直接賦值給下面的DTO

	// 組合 DTO
	res := &PostResponse{
		PostID:         post.PostID,
		PictureURL:     pics,
		Author:         strings.TrimSpace(author.LastName + author.FirstName),
		AuthorID:       author.UserID,
		AuthorPic:      author.ProfilePicture,
		Title:          post.Title,
		Timestamp:      post.CreatedAt.Format("2006-01-02 15:04:05"), //!!如果用.string()會有包含時區 我目前先不管時區
		Content:        post.Content,
		LikeNumber:     post.LikeNumber,
		BookmarkNumber: post.SaveNumber,
		Tags:           post.Tags,
		ILikeThis:      ilikethis,
		ISaveThis:      isavethis,
		Latitude:		post.Latitude,
		Longitude:		post.Longitude,
		Address:		post.Address,
	}


	return res, nil
}
func (s *PostService) CreatePost(
	userID string,
	title string,
	address string,
	latitude float64,
	longitude float64,
	content string,
	tags datatypes.JSON,
	files []*multipart.FileHeader,
) (*models.Post, error) {
	pictureURLs, err := s.uploadPostPicturesToSupabase(context.Background(), userID, files)
	if err != nil {
		return nil, err
	}

	pictureJSON, err := json.Marshal(pictureURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to encode pictures: %w", err)
	}

	city, district := parseCityDistrictFromAddress(address)
	if !isAllowedCityDistrict(city, district) {
		return nil, fmt.Errorf("failed to get the city and zone: %w", err)
	}

	post := models.Post{
		UserID:     userID,
		Title:      title,
		Picture:    datatypes.JSON(pictureJSON),
		Address:    address,
		City:       city,
		District:   district,
		Latitude:   latitude,
		Longitude:  longitude,
		CreatedAt:  time.Now().Truncate(time.Second),
		Content:    content,
		LikeNumber: 0,
		SaveNumber: 0,
		Tags:       tags,
	}

	// Create 處理所有一般欄位（含 deleted_at 自動 NULL）
	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	// 單獨更新 PostGIS location 欄位
	if err := s.db.Exec(`
		UPDATE posts SET location = ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography
		WHERE post_id = ?
	`, longitude, latitude, post.PostID).Error; err != nil {
		return nil, err
	}

	return &post, nil
}
func parseCityDistrictFromAddress(address string) (string, string) {
	normalized := strings.TrimSpace(address)
	normalized = strings.TrimPrefix(normalized, "台灣")
	normalized = strings.TrimPrefix(normalized, "臺灣")

	city := ""
	district := ""

	if cityEnd := strings.IndexRune(normalized, '市'); cityEnd != -1 {
		city = strings.TrimSpace(normalized[:cityEnd+len("市")])
		city = strings.ReplaceAll(city, "臺", "台")
		rest := normalized[cityEnd+len("市"):]

		if districtEnd := strings.IndexRune(rest, '區'); districtEnd != -1 {
			district = strings.TrimSpace(rest[:districtEnd+len("區")])
		}
	}

	return city, district
}

func isAllowedCityDistrict(city string, district string) bool {
	if city == "" || district == "" {
		return false
	}

	districtSet, ok := allowedDistrictsByCity[city]
	if !ok {
		return false
	}
	_, exists := districtSet[district]
	return exists
}

func (s *PostService) uploadPostPicturesToSupabase(ctx context.Context, userID string, files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return []string{}, nil
	}

	supabaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is not configured")
	}

	supabaseKey := strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	if supabaseKey == "" {
		// fallback：沿用你目前 workflow 使用的 key 名稱
		supabaseKey = strings.TrimSpace(os.Getenv("SUPABASE_KEY"))
	}
	if supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY/SUPABASE_KEY is not configured")
	}

	bucketName := strings.TrimSpace(os.Getenv("SUPABASE_STORAGE_BUCKET"))
	if bucketName == "" {
		return nil, fmt.Errorf("SUPABASE_STORAGE_BUCKET is not configured")
	}

	userSegment := sanitizePathSegment(userID)
	pictureURLs := make([]string, 0, len(files))
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		objectName := fmt.Sprintf("posts/%s/%s%s", userSegment, uuid.NewString(), ext)

		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open upload file: %w", err)
		}
		contentBytes, err := io.ReadAll(src)
		_ = src.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read upload file: %w", err)
		}

		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, objectName)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(contentBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to build supabase upload request: %w", err)
		}
		contentType := file.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		req.Header.Set("Authorization", "Bearer "+supabaseKey)
		req.Header.Set("apikey", supabaseKey)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("x-upsert", "true")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image to supabase storage: %w", err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("supabase upload failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		pictureURLs = append(pictureURLs, fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, objectName))
	}

	return pictureURLs, nil
}

func sanitizePathSegment(input string) string {
	if input == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}

func (ps *PostService) DeletePost(postID int) error { //我已經用了軟刪除所以可以直接刪
	if err := ps.db.Where("post_id = ?", postID).Delete(&models.Post{}).Error; err != nil {
		return err
	}
	return nil
}

func (ps *PostService) LikePostByID(postID int, userID string) error {
	like := models.PostLike{
		UserID: userID,
		PostID: postID,
	}

	if err := ps.db.Create(&like).Error; err != nil {
		return err
	}

	if err := ps.db.Model(&models.Post{}).Where("post_id = ?", postID).UpdateColumn("like_number", gorm.Expr("like_number + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (ps *PostService) UnlikePostByID(postID int, userID string) error {
	if err := ps.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.PostLike{}).Error; err != nil {
		return err
	}

	if err := ps.db.Model(&models.Post{}).Where("post_id = ?", postID).UpdateColumn("like_number", gorm.Expr("like_number - ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (ps *PostService) SavePostByID(postID int, userID string) error {
	save := models.PostSave{
		UserID: userID,
		PostID: postID,
	}

	if err := ps.db.Create(&save).Error; err != nil {
		return err
	}

	if err := ps.db.Model(&models.Post{}).Where("post_id = ?", postID).UpdateColumn("save_number", gorm.Expr("save_number + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (ps *PostService) GetSavePost(userID string, orderDir string) ([]SavePostPageResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return []SavePostPageResponse{}, nil
	}

	type savedPostRow struct {
		PostID    int            `gorm:"column:post_id"`
		Title     string         `gorm:"column:title"`
		CreatedAt time.Time      `gorm:"column:created_at"`
		Picture   datatypes.JSON `gorm:"column:picture"`
		Content   string         `gorm:"column:content"`
		City      string         `gorm:"column:city"`
		District  string         `gorm:"column:district"`
	}

	var rows []savedPostRow
	err := ps.db.Table("post_saves AS ps").
		Select("p.post_id, p.title, p.created_at, p.picture, p.content, p.city, p.district").
		Joins("JOIN posts AS p ON p.post_id = ps.post_id").
		Where("ps.user_id = ?", userID).
		Order("ps.created_at " + orderDir).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	res := make([]SavePostPageResponse, 0, len(rows))
	for _, row := range rows {
		var pics []string
		if len(row.Picture) > 0 {
			if err := json.Unmarshal(row.Picture, &pics); err != nil {
				return nil, err
			}
		}

		res = append(res, SavePostPageResponse{
			PostID:     row.PostID,
			Title:      row.Title,
			Timestamp:  row.CreatedAt.Format("2006-01-02 15:04:05"),
			PictureURL: pics,
			Content:    row.Content,
			City:       row.City,
			District:   row.District,
		})
	}

	return res, nil
}

func (s *PostService) UnsavePostByID(postID int, userID string) error {
	if err := s.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.PostSave{}).Error; err != nil {
		return err
	}

	if err := s.db.Model(&models.Post{}).Where("post_id = ?", postID).UpdateColumn("save_number", gorm.Expr("save_number - ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (s *PostService) ReportPostByID(postID int, reportReason string) error {
	return nil 
}
