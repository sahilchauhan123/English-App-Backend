package userservice

import (
	"fmt"
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

	url, err := s3.UploadFile(newFileName, src)
	if err != nil {
		return "", err

	}

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
