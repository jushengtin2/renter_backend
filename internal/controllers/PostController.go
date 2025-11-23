package controllers

import (
	"fmt"
	"net/http"
	"renter_backend/internal/services"

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

func (pc *PostController) GetPostForMainPage(c *gin.Context) { //*gin.Context = 一次 HTTP 請求的上下文物件。所以只有controller知道context
	post_filter := c.Query("filter") // 取得查詢參數 filter
    userID, _ := c.Get("clerk_user_id") 
    userIDStr, ok := userID.(string)
	if !ok {
		// if the user id is missing or not a string, use empty string (or handle as needed)
		userIDStr = ""
	}
	posts, err := pc.post_service.GetPostForMainPage(post_filter, userIDStr)
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

//發文需要驗證jwt
func (pc *PostController) CreatePost(c *gin.Context) {
    userID := c.GetString("clerk_user_id") // 從 context 取得 user id
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    fmt.Println("userID from token:", userID)
    
	type PostRequest struct {
    PostTitle     string   `json:"post_title" binding:"required"`
    PostPicture   datatypes.JSON   `jsonb:"post_picture"`
    PostLocation  string   `json:"post_location" binding:"required"`
    PostContent   string   `json:"post_content" binding:"required"`
    PostTag       []string `json:"post_tag"`
	}

    var para PostRequest
    if err := c.ShouldBindJSON(&para); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 呼叫 Service 建立文章
    // if _, err := pc.post_service.CreatePost(
    //     userID,
    //     para.PostTitle,
    //     para.PostPicture,
    //     para.PostLocation,
    //     para.PostContent,
    //     para.PostTag,
    // ); err != nil {
    //     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    //     return
    // }

    c.JSON(http.StatusCreated, gin.H{"message": "post created successfully"})
}

func(pc *PostController) DeletePost(c *gin.Context) {
    type DeletePostRequest struct {
        PostID int `json:"post_id" binding:"required"`
    }
    var para DeletePostRequest
    if err := c.ShouldBindJSON(&para); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    fmt.Println("postID:", para.PostID)

    // if err := pc.post_service.DeletePost(para.PostID); err != nil {
    //     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    //     return
    // }
    c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})


}
