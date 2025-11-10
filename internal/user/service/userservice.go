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

func GetCallHistory(userId int64, db storage.Storage, timestamp time.Time) ([]types.CallRecord, error) {
	history, err := db.GetCallHistory(userId, timestamp)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func DeleteAccount(userId int64, db storage.Storage) error {
	err := db.DeleteAccount(userId)
	if err != nil {
		return err
	}
	return nil
}

func DeletePicture(userId int64, imageUrl string, db storage.Storage) error {
	err := db.DeletePicture(userId, imageUrl)
	if err != nil {
		return err
	}
	return nil
}

func BlockUser(userid int64, blockUserId int64, db storage.Storage) error {
	err := db.BlockUser(userid, blockUserId)
	if err != nil {
		return err
	}
	return nil
}

func GetLeaderboard(db storage.Storage, duration string) ([]types.LeaderboardEntry, error) {
	leaderboard, err := db.GetLeaderboard(duration)
	if err != nil {
		return nil, err
	}
	return leaderboard, nil
}

func GetAiCharacters() []types.AiCharacters {
	characters := []types.AiCharacters{
		{
			Id:              0,
			Name:            "Flirty Girl",
			PicUrl:          "https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/wxnetp7ly2rprc5nxks9.png",
			Description:     "Talk with her to forget your pain",
			BackgroundColor: "#F7AEB0",
		},
		{
			Id:              1,
			Name:            "English Teacher",
			PicUrl:          "https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/a7uscfnbccswajyfnvnk.png",
			Description:     "He will help to improve your English",
			BackgroundColor: "#DA8F7A",
		},
		{
			Id:              2,
			Name:            "Female Crush",
			PicUrl:          "https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/c6yx5yjtvolnu05epykf.png",
			Description:     "Talk with her Sure you will in love with her",
			BackgroundColor: "#9CCABD",
		},
		{
			Id:              3,
			Name:            "Motivator",
			PicUrl:          "https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/ovdqut37etazxo4935ym.png",
			Description:     "Helps in Motivation and overcoming your fear",
			BackgroundColor: "#AAC2CC",
		},
		{
			Id:              4,
			Name:            "Lawyer",
			PicUrl:          "https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/tvssabzep2mxy8kqdtvf.png",
			Description:     "Talk and get some advice for your cases",
			BackgroundColor: "#DDDDD8",
		},
	}
	// DONT FORGET TO ADD THIS TO LIST
	// https://pub-ab9111ea2a9741bba809ddf443459af8.r2.dev/Group%2022%20(1).png

	return characters
}

func SaveCallFeedback(feedback types.CallFeedbackRequest, db storage.Storage) error {

	err := db.SaveCallFeedback(feedback)
	if err != nil {
		return err
	}
	fmt.Println("Feedback saved successfully")
	return nil
}

func GetCallFeedback(callID string, db storage.Storage) ([]types.CallFeedbackResponse, error) {
	feedbacks, err := db.GetCallFeedback(callID)
	if err != nil {
		return nil, err
	}
	return feedbacks, nil
}
