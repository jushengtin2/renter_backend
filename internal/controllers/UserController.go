package controllers

import (
	"fmt"
	"renter_backend/internal/services"

	"github.com/gin-gonic/gin"
)


func NewUserController(s *services.UserService) *UserController {
	return &UserController{user_service: s}
}

type UserController struct {
	// 這邊可以放很多指標物件 例如下面這個
	user_service *services.UserService //自己命名的user_service 然後用指標指向Service層已經創好的東西就不用每次請求都重新new東西， 意思是UserController依賴UserService，它不自己處理商業邏輯，而是交給 Service。
	// 就像是我的Controller在創立的時候，裡面已經包了一個service，讓我在下面宣告func的時候可以func (uc *UserController)
}

type ClerkAPI_CreateUser_Payload struct {
	Data struct {
		ID                    string  `json:"id"`
		FirstName             *string `json:"first_name"`
		LastName              *string `json:"last_name"`
		ImageURL              *string `json:"image_url"`
		PrimaryEmailAddressID string  `json:"primary_email_address_id"` //因爲clerk的API裡面email是放在一個陣列裡面（然後裡面還有很多email object)，然後primary_email_address_id是告訴你哪一個email是主要的，所以我們要根據這個ID去找email陣列裡面對應的email地址
		EmailAddresses        []struct {
			ID           string `json:"id"`
			EmailAddress string `json:"email_address"`
		} `json:"email_addresses"`
	} `json:"data"`
	Type string `json:"type"` // 這是哪個狀態下發的webhook 比如 "user.created" "user.deleted" "user.updated"
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func extractPrimaryEmail(payload ClerkAPI_CreateUser_Payload) string {
	if payload.Data.PrimaryEmailAddressID != "" {
		for _, e := range payload.Data.EmailAddresses {
			if e.ID == payload.Data.PrimaryEmailAddressID {
				fmt.Println("Primary email found:", e.EmailAddress)
				return e.EmailAddress
			}
		}
	}

	if len(payload.Data.EmailAddresses) > 0 {
		fmt.Println("Primary email not found, using first email in list:", payload.Data.EmailAddresses[0].EmailAddress)
		return payload.Data.EmailAddresses[0].EmailAddress
	}
	fmt.Println("No email addresses found for user:", payload.Data.ID)
	return ""
}

func (uc *UserController) GetUserProfile(c *gin.Context) { //這個是會接收clerk傳來的新user info
	fmt.Println("GetUserProfile called")
	var payload ClerkAPI_CreateUser_Payload
	if err:= c.ShouldBindJSON(&payload); err != nil {
		fmt.Println("Error parsing JSON:", err)
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	switch payload.Type {
	case "user.created":

		//檢查裡面是否都有值
		if payload.Data.ID == "" {
			c.JSON(400, gin.H{"error": "Missing user ID"})
			return
		}
		firstName := stringValue(payload.Data.FirstName)
		lastName := stringValue(payload.Data.LastName)
		if firstName == "" && lastName == "" {
			c.JSON(400, gin.H{"error": "Missing user name (full name is empty)"})
			return
		}

		emailStr := extractPrimaryEmail(payload)
		if _, err := uc.user_service.GetUserProfile("created", payload.Data.ID, firstName, lastName, emailStr, stringValue(payload.Data.ImageURL)); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	case "user.updated":
		if payload.Data.ID == "" {
			c.JSON(400, gin.H{"error": "Missing user ID"})
			return
		}
		firstName := stringValue(payload.Data.FirstName)
		lastName := stringValue(payload.Data.LastName)
		if firstName == "" && lastName == "" {
			c.JSON(400, gin.H{"error": "Missing user name (full name is empty)"})
			return
		}

		emailStr := extractPrimaryEmail(payload)
		if _, err := uc.user_service.GetUserProfile("updated", payload.Data.ID, firstName, lastName, emailStr, stringValue(payload.Data.ImageURL)); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	case "user.deleted":

		if _, err := uc.user_service.GetUserProfile("deleted", payload.Data.ID, "", "", "", ""); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(400, gin.H{"error": "Unknown webhook type"})
		return
	}

	c.JSON(200, gin.H{"message": "User profile processed (from clerk webhook) successfully"})
	
}
