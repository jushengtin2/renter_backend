package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"renter_backend/internal/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func NewPostController(s *services.PostService) *PostController { //*PostController是回傳型別
	// 依賴注入(Dependency Injection)模式，先有service再用controller去對到他，所以保證Controller建立時，一定有Service注入。
	// 好測試：你可以傳一個假 Service 進去 Controller 做單元測試。
	// 低耦合：未來換成 PostServiceV2 也很容易，只要在建構時替換。
	return &PostController{post_service: s}
}

type PostController struct {
	// 這邊可以放很多指標物件 例如下面這個
	post_service *services.PostService //自己命名的post_service 然後用指標指向Service層已經創好的東西就不用每次請求都重新new東西， 意思是PostController依賴PostService，它不自己處理商業邏輯，而是交給 Service。
	// 就像是我的Controller在創立的時候，裡面已經包了一個service，讓我在下面宣告func的時候可以func (pc *PostController)
}

//(pc * PostController) 代表 這個方法是綁定在 *PostController 這個 struct 上的 (指標操作！！)
//當你呼叫 controller.GetPostForMainPage(c) 時，Go 編譯器會自動把 controller 傳進去，對應到這裡的 pc
//Go 的 pc 就等於 Java 的 this、Python 的 self。只是 Go 沒有固定名字，你可以自己取名
//語法設計：Go 不用 class，而是用 struct + method receiver 來模擬 OOP。
//彈性：接收者可以是「值型別」( PostController ) 或「指標型別」( *PostController )。
//用值 → 方法操作的是副本。
//用指標 → 方法操作的是原本的 struct（常用）。

func (pc *PostController) GetPostForMainPage(c *gin.Context) { 
	search_text := c.Query("search_text")
	sort := c.Query("sort") // 熱門最新最舊
	area := c.Query("area") 
	tag := c.Query("tags")
	pageValue := c.Query("page")
	pageSizeValue := c.Query("page_size")
	minLatValue, hasMinLat := c.GetQuery("minLat")
	maxLatValue, hasMaxLat := c.GetQuery("maxLat")
	minLngValue, hasMinLng := c.GetQuery("minLng")
	maxLngValue, hasMaxLng := c.GetQuery("maxLng")

	userID, _ := c.Get("clerk_user_id")
	userIDStr, ok := userID.(string)
	if !ok {
		// if the user id is missing or not a string, use empty string (or handle as needed)
		userIDStr = ""
	}
	if sort == "" {
		sort = "hot" // default sort
	}
	if area == "" {
		area = "all" // default area
	}
	var tags []string
	if tag != "" {
		tags = strings.Split(tag, ",")
	}

	fmt.Println("hasMinLat: ",hasMinLat)
	fmt.Println("search_text: ",search_text)


	page := 1 //我後端預設是回傳"目前是第一頁" 但也可以從前端傳來改變這個值
	if pageValue != "" {
		val, err := strconv.Atoi(pageValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
			return
		}
		page = val
	}

	pageSize := 10 //我後端預設是回傳10筆 但也可以從前端傳來改變這個值
	if pageSizeValue != "" {
		val, err := strconv.Atoi(pageSizeValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
			return
		}
		pageSize = val
	}

	minLat := math.NaN()
	maxLat := math.NaN()
	minLng := math.NaN()
	maxLng := math.NaN()

	if hasMinLat || hasMaxLat || hasMinLng || hasMaxLng {
		if !(hasMinLat && hasMaxLat && hasMinLng && hasMaxLng) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "minLat, maxLat, minLng, maxLng must be provided together"})
			return
		}

		var err error
		minLat, err = strconv.ParseFloat(minLatValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid minLat"})
			return
		}
		maxLat, err = strconv.ParseFloat(maxLatValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid maxLat"})
			return
		}
		minLng, err = strconv.ParseFloat(minLngValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid minLng"})
			return
		}
		maxLng, err = strconv.ParseFloat(maxLngValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid maxLng"})
			return
		}

		// ！！！！當使用者選地圖時 不會用下拉式的行政區篩選！！！！ 主要是保底用 我會在前端寫只要使用者拖曳了地圖就把下拉式的行政區選擇結果值清掉，讓他不會兩個都選
		area = ""
	}

	posts, err := pc.post_service.GetPostForMainPage(sort, search_text, area, tags, userIDStr, page, pageSize, minLat, maxLat, minLng, maxLng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts) // c.JSON代表回傳json格式
}

func (pc *PostController) GetPostByID(c *gin.Context) {
	postID := c.Param("postid")
	userID, _ := c.Get("clerk_user_id")

	userIDStr, ok := userID.(string)
	if !ok {
		// if the user id is missing or not a string, use empty string (or handle as needed)
		userIDStr = ""
	}

	res, err := pc.post_service.GetPostByID(postID, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// 發文需要驗證jwt
func (pc *PostController) CreatePost(c *gin.Context) {
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	title := c.PostForm("post_title")
	address := c.PostForm("post_address")
	content := c.PostForm("post_content")
	latStr := c.PostForm("post_latitude")
	lngStr := c.PostForm("post_longitude")
	tagStr := c.PostForm("post_tag")

	//因為傳來的是formdata所以要解析

	if title == "" || address == "" || content == "" || latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required form fields"})
		return
	}

	if !strings.Contains(address, "台北市") && !strings.Contains(address, "臺北市") && !strings.Contains(address, "新北市") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is out of taipei area"})
		return
	}

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_latitude"})
		return
	}

	longitude, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_longitude"})
		return
	}

	parsedTags := make([]string, 0)
	if tagStr != "" {
		if err := json.Unmarshal([]byte(tagStr), &parsedTags); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "post_tag must be a JSON array string"})
			return
		}
	}
	tagsJSON, err := json.Marshal(parsedTags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode post_tag"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart/form-data"})
		return
	}

	pic_files := form.File["post_picture"]

	// 呼叫 Service 建立文章
	if _, err := pc.post_service.CreatePost(
		userID,
		title,
		address,
		latitude,
		longitude,
		content,
		datatypes.JSON(tagsJSON),
		pic_files,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully"})
}

func (pc *PostController) DeletePost(c *gin.Context) {
	postID := c.Param("postid")
	if postID == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "postid is required"})
	}
	postIDint , err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	
	if err := pc.post_service.DeletePost(postIDint); err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	    return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})

}

func (pc *PostController) LikePostByID(c *gin.Context) {
	postID := c.Param("postid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := pc.post_service.LikePostByID(postIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (pc *PostController) UnlikePostByID(c *gin.Context) {
	postID := c.Param("postid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := pc.post_service.UnlikePostByID(postIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (pc *PostController) SavePostByID(c *gin.Context) {
	postID := c.Param("postid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := pc.post_service.SavePostByID(postIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (pc *PostController) GetSavePost(c *gin.Context){
	userID := c.GetString("clerk_user_id") 
	sort := c.Query("sort")
	var orderDir string 

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or no user info"})
		return
	}
	if sort == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "no sort info"})
		return
	}
	if sort != "new" && sort != "old" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort info"})
		return
	}

	if sort == "new"{
		orderDir = "DESC"
	} else{
		orderDir = "ASC"
	}


	res, err := pc.post_service.GetSavePost(userID, orderDir)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)

}

func (pc *PostController) UnSavePostByID(c *gin.Context) {
	postID := c.Param("postid")
	userID := c.GetString("clerk_user_id") // 已經從middleware context 取得 user id

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println("userID from token:", userID)

	if err := pc.post_service.UnsavePostByID(postIDInt, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post liked successfully"})

}

func (pc *PostController) ReportPostByID(c *gin.Context) {
	postID := c.Param("postid")
	reportReason := c.Query("reason") 
	if reportReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report reason is required"})
		return
	}
	postID_int, err := strconv.Atoi(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postid"})
		return
	}

	if err := pc.post_service.ReportPostByID(postID_int, reportReason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "report successfully"})
}
