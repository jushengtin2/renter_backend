package controllers

import (
	"net/http"
	"renter_backend/internal/services"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	// 這邊可以放很多指標物件 例如下面這個
    post_service *services.PostService //自己命名的post_service 然後用指標指向Service層已經創好的東西就不用每次請求都重新new東西， 意思是PostController依賴PostService，它不自己處理商業邏輯，而是交給 Service。
}

func NewPostController(s *services.PostService) *PostController { //*PostController是回傳型別
	// 依賴注入(Dependency Injection)模式，先有service再用controller去對到他，所以保證Controller建立時，一定有Service注入。
	// 好測試：你可以傳一個假 Service 進去 Controller 做單元測試。
	// 低耦合：未來換成 PostServiceV2 也很容易，只要在建構時替換。
    return &PostController{post_service: s}
}

//(pc * PostController) 代表 這個方法是綁定在 *PostController 這個 struct 上的 (指標操作！！)
//當你呼叫 controller.GetPostForMainPage(c) 時，Go 編譯器會自動把 controller 傳進去，對應到這裡的 pc
//Go 的 pc 就等於 Java 的 this、Python 的 self。只是 Go 沒有固定名字，你可以自己取名
//語法設計：Go 不用 class，而是用 struct + method receiver 來模擬 OOP。
//彈性：接收者可以是「值型別」( PostController ) 或「指標型別」( *PostController )。
//用值 → 方法操作的是副本。
//用指標 → 方法操作的是原本的 struct（常用）。
func (pc *PostController) GetPostForMainPage (c *gin.Context) { //*gin.Context = 一次 HTTP 請求的上下文物件。所以只有controller知道context
	post_filter := c.Query("filter") // 取得查詢參數 filter
	posts, err := pc.post_service.GetPostForMainPage(post_filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts) // c.JSON代表回傳json格式
}

func (pc *PostController) GetPostByID (c *gin.Context) { 
	postID := c.Param("postid")
	posts, err := pc.post_service.GetPostByID(postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts) 
}