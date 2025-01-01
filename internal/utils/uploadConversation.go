package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"smart-chat/internal/store"
	"time"

	// Import the AWS SDK packages
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	bucketName = "indtrav-smart-chat-conversations"
	awsRegion  = "ap-south-1"
)

func PushConversationsToS3() {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion), // e.g., "us-west-2"
	})
	// sess, err := session.NewSession(&aws.Config{
	// 	Region:      aws.String(awsRegion),
	// 	Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	// })
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return
	}

	uploader := s3manager.NewUploader(sess)

	conversations, err := store.FetchAllConversations()
	if err != nil {
		log.Printf("Error fetching conversations: %v", err)
		return
	}

	currentTime := time.Now()
	for _, conversation := range conversations {
		if conversation.Pushed {
			continue
		}

		if currentTime.Sub(conversation.AccessTokenExpireTime) <= 2*time.Hour {
			continue
		}

		err := uploadToS3(uploader, conversation)
		if err != nil {
			log.Printf("Error uploading conversation to S3: %v", err)
			continue
		}

		err = store.MarkConversationAsPushed(conversation.AccessToken)
		if err != nil {
			log.Printf("Error marking conversation as pushed: %v", err)
		}

		// Delete the conversation and check for errors
		err = store.DeleteConversation(conversation.AccessToken)
		if err != nil {
			log.Printf("Error deleting conversation: %v", err)
		}

	}
}

func uploadToS3(uploader *s3manager.Uploader, conversation store.Conversation) error {
	// Marshal the conversation to JSON
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}

	// Construct the file name using the Username and AccessToken
	fileName := conversation.UserName + "-conversation-" + conversation.AccessToken

	// Define the S3 upload input
	input := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(data),
	}

	// Perform the upload
	_, err = uploader.Upload(input)
	return err
}
