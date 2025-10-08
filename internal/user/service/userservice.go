package userservice

import (
	"fmt"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"github/english-app/storage/s3"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

func HandleUploadImage(userID int64, file *multipart.FileHeader, db storage.Storage, s3 *s3.Repo, ext string) (string, error) {
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	pictureLength, err := db.CheckPictureLength(userID)
	if err != nil {
		return "", err
	}
	if pictureLength >= 4 {
		return "", fmt.Errorf("maximum 4 pictures are allowed")
	}
	url, err := s3.UploadFile(newFileName, src)
	if err != nil {
		return "", err
	}
	err = db.InsertPicture(userID, url)
	if err != nil {
		return "", err
	}
	fmt.Println("Uploaded file URL:", url)
	return url, nil
}

func DetectFileType(file *multipart.FileHeader) (string, error) {
	// open file content (without saving to disk)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// detect content type
	detectedType := http.DetectContentType(buffer)
	fmt.Println("Detected Content-Type:", detectedType)

	// reset file reader to beginning
	src.Seek(0, io.SeekStart)

	return detectedType, nil
}

func GetProfile(userId int64, db storage.Storage) (types.User, error) {
	profile, err := db.GetProfile(userId)
	if err != nil {
		return types.User{}, err
	}
	return profile, nil
}

func GetCallHistory(userId int64, db storage.Storage) ([]types.CallHistory, error) {
	history, err := db.GetCallHistory(userId)
	if err != nil {
		return nil, err
	}
	return history, nil
}
