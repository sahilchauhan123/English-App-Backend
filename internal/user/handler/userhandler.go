package userhandler

import (
	"fmt"
	"github/english-app/internal/response"
	userservice "github/english-app/internal/user/service"
	"github/english-app/storage"
	"github/english-app/storage/s3"
	"net/http"
	"path/filepath"
	"strings"

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
