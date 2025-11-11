package userhandler

import (
	"fmt"
	"github/english-app/internal/response"
	"github/english-app/internal/types"
	userservice "github/english-app/internal/user/service"
	"github/english-app/storage"
	"github/english-app/storage/s3"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// func UploadImageHandler(db storage.Storage, s3 *s3.Repo) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userid := c.MustGet("user_id").(int64)

// 		// get file
// 		file, err := c.FormFile("image")
// 		if err != nil {
// 			response.Failed(c, http.StatusBadRequest, "No file received")
// 			return
// 		}

// 		fmt.Println("file Name:", file.Filename)

// 		// extension (lowercased)
// 		ext := strings.ToLower(filepath.Ext(file.Filename))
// 		f, _ := os.Open(file.Filename)
// 		buffer := make([]byte, 512)
// 		f.Read(buffer)
// 		http.DetectContentType(buffer)
// 		// mime type from header
// 		contentType := file.Header.Get("Content-Type")

// 		// validate
// 		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
// 			if contentType == "image/jpeg" {
// 				ext = ".jpg"
// 			} else if contentType == "image/png" {
// 				ext = ".png"
// 			} else {
// 				response.Failed(c, http.StatusBadRequest, "Invalid file type")
// 				return
// 			}
// 		}

// 		// handle upload
// 		url, err := userservice.HandleUploadImage(userid, file, db, s3, ext)
// 		if err != nil {
// 			response.Failed(c, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		// success response
// 		response.Success(c, map[string]any{
// 			"success": true,
// 			"url":     url,
// 		})
// 	}
// }

func UploadImageHandler(db storage.Storage, s3 *s3.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.MustGet("user_id").(int64)

		// get uploaded file
		file, err := c.FormFile("image")
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "No file received")
			return
		}

		fmt.Println("file Name:", file.Filename)
		detectedType, err := userservice.DetectFileType(file)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, "Error detecting file type")
			return
		}
		// extension (lowercased)
		ext := strings.ToLower(filepath.Ext(file.Filename))

		// validate by ext OR detected type
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			if detectedType == "image/jpeg" {
				ext = ".jpg"
			} else if detectedType == "image/png" {
				ext = ".png"
			} else {
				response.Failed(c, http.StatusBadRequest, "Invalid file type")
				return
			}
		}

		// handle upload (pass ext)
		url, err := userservice.HandleUploadImage(userid, file, db, s3, ext)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"url":     url,
		})
	}
}

func GetProfileHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.MustGet("user_id").(int64)
		// app := notifications.InitializeAppWithServiceAccount()
		// notifications.SendToToken(app)
		profile, err := userservice.GetProfile(userid, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"profile": profile,
		})
	}
}

func GetOtherUserProfileHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.Param("id")
		userId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
		}
		//checking if same user requested the other user profile
		// if userId == c.MustGet("user_id").(int64) {
		// 	response.Failed(c, http.StatusInternalServerError, "trying to request own id")
		// }

		profile, err := userservice.GetProfile(userId, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"profile": profile,
		})
	}
}

func GetCallHistoryHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.MustGet("user_id").(int64)
		timestampStr := c.Query("timestamp")
		timestampStr = strings.ReplaceAll(timestampStr, " ", string("T"))

		timestamp, err := time.Parse("2006-01-02T15:04:05", timestampStr)
		if err != nil && timestampStr != "" {
			response.Failed(c, http.StatusBadRequest, "Invalid timestamp format")
			return
		}

		history, err := userservice.GetCallHistory(userid, db, timestamp)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"history": history,
		})
	}
}

func DeleteAccountHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.MustGet("user_id").(int64)

		err := userservice.DeleteAccount(userid, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"message": "Account deleted successfully",
		})
	}
}

func DeletePictureHandler(db storage.Storage, s3 *s3.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.MustGet("user_id").(int64)
		picUrl := c.Query("url")
		if picUrl == "" {
			response.Failed(c, http.StatusBadRequest, "No picture URL provided")
			return
		}

		err := userservice.DeletePicture(userid, picUrl, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"message": "Picture deleted successfully",
		})
	}
}

func BlockUserHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {

		value := c.Query("id")
		blockUserId, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
			return
		}

		userid := c.MustGet("user_id").(int64)

		err = db.BlockUser(userid, blockUserId)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success": true,
			"message": "User blocked successfully",
		})
	}
}

func LeaderboardHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		Duration := c.Query("duration") // daily, weekly, monthly, alltime
		if Duration == "" {
			response.Failed(c, http.StatusBadRequest, "Duration is required")
			return
		}
		leaderboard, err := userservice.GetLeaderboard(db, Duration)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, map[string]any{
			"success":     true,
			"leaderboard": leaderboard,
		})
	}
}

func AiCharactersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		characters := userservice.GetAiCharacters()
		response.Success(c, map[string]any{
			"success":    true,
			"characters": characters,
		})
	}
}

func SubmitCallFeedbackHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var feedback types.CallFeedbackRequest
		//validate
		if err := c.ShouldBindJSON(&feedback); err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
			return
		}

		// Get rater ID from context (set by auth middleware)
		raterID, _ := c.Get("user_id")
		feedback.RaterUserID = raterID.(int64)

		err := userservice.SaveCallFeedback(feedback, db)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
			return
		}
		fmt.Println("done broooooo")
		response.Success(c, map[string]any{
			"success": true,
			"message": "Feedback submitted successfully",
		})
	}
}

func GetCallFeedbackHandler(db storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {

		userId := c.MustGet("user_id").(int64)

		feedbacks, err := userservice.GetCallFeedback(userId, db)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Success(c, feedbacks)
	}
}
