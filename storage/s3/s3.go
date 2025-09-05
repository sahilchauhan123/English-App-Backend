package s3

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repo struct {
	s3Client          *s3.Client
	s3PresignedClient *s3.PresignClient
}

func NewS3Client(accountID, accessKey, secretKey, bucketRegion string) *Repo {
	resolver := s3.EndpointResolverFromURL(
		"https://" + accountID + ".r2.cloudflarestorage.com",
	)

	options := s3.Options{
		Region:           bucketRegion, // use "auto" for R2
		Credentials:      aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		EndpointResolver: resolver,
	}

	client := s3.New(options)
	presignClient := s3.NewPresignClient(client)

	return &Repo{
		s3Client:          client,
		s3PresignedClient: presignClient,
	}
}

func (repo Repo) PutObject(bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := repo.s3PresignedClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to put %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}

func (repo Repo) DeleteObject(bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := repo.s3PresignedClient.PresignDeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to delete %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}

	return request, err
}

// func (repo Repo) UploadFile(file image.Image, url string) error {
// 	var buf bytes.Buffer
// 	err := jpeg.Encode(&buf, file, nil)
// 	if err != nil {
// 		return nil
// 	}
// 	body := io.Reader(&buf)
// 	request, err := http.NewRequest(http.MethodPut, url, body)
// 	if err != nil {
// 		return err
// 	}

// 	request.Header.Set("Content-Type", "image/jpeg")

// 	client := &http.Client{}
// 	resp, err := client.Do(request)
// 	if err != nil {
// 		log.Println("Error sending request:", err)
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return err
// }

func (repo Repo) DeleteFile(url string) error {
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	log.Printf("Delete Request Response status code: %v", response.StatusCode)
	return err
}

func (repo Repo) UploadFile(fileName string, file multipart.File) (string, error) {
	_, err := repo.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("R2_BUCKET_NAME")),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://%s/%s/%s",
		os.Getenv("R2_ACCOUNT_ID")+".r2.cloudflarestorage.com",
		os.Getenv("R2_BUCKET_NAME"),
		fileName,
	)
	return url, nil
}
