package config

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/gin-gonic/gin"
)

type Config struct {
	OpenAIKey              string
	FAST2SMS_API_KEY       string
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	Email                  string
	EmailPassword          string
	SecretToken            string
	IndianTeavellersURL    string
	NotificationServiceURL string
	SlackNotificationURL   string
	SlackAlertURL          string
}

func Load() *Config {
	config := &Config{
		OpenAIKey:              "default-openai-key",
		FAST2SMS_API_KEY:       "default-fast2sms-api-key",
		DBHost:                 "localhost",
		DBPort:                 "3306",
		DBUser:                 "root",
		DBPassword:             "password",
		DBName:                 "smart_chat",
		Email:                  "test@email.com",
		EmailPassword:          "test_pwd",
		SecretToken:            "secret_token",
		IndianTeavellersURL:    "http://127.0.0.1:8000",
		NotificationServiceURL: "http://127.0.0.1:8001",
		SlackNotificationURL:   "https://hooks.slack.com/services/xx",
		SlackAlertURL:          "https://hooks.slack.com/services/xx",
	}
	if os.Getenv("SMART_CHAT_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("ap-south-1"),
		})
		if err != nil {
			log.Fatalf("Failed to create AWS session: %v", err)
		}

		ssmSvc := ssm.New(sess)

		getParameter := func(name string) string {
			withDecryption := true
			param, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
				Name:           &name,
				WithDecryption: &withDecryption,
			})
			if err != nil {
				log.Fatalf("Failed to get parameter: %v", err)
			}
			return *param.Parameter.Value
		}

		config.OpenAIKey = getParameter("OpenAIKey")
		config.FAST2SMS_API_KEY = getParameter("FAST2SMS_API_KEY")
		config.DBHost = getParameter("SmartChatDBHost")
		config.DBPort = getParameter("SmartChatDBPort")
		config.DBUser = getParameter("SmartChatDBUser")
		config.DBPassword = getParameter("SmartChatDBPassword")
		config.DBName = getParameter("SmartChatDBName")
		config.Email = getParameter("EmailAddress")
		config.EmailPassword = getParameter("EmailPassword")
		config.SecretToken = getParameter("WASecretToken")
		config.IndianTeavellersURL = getParameter("IndianTeavellersURL")
		config.NotificationServiceURL = getParameter("NotificationServiceURL")
		config.SlackNotificationURL = getParameter("SLACK_NOTIFICATION_URL")
		config.SlackAlertURL = getParameter("SLACK_ALERT_URL")
	} else {
		gin.SetMode(gin.DebugMode)
		return &Config{
			OpenAIKey:              "sk-xxxxx",
			FAST2SMS_API_KEY:       "xxxxxxx",
			DBHost:                 "localhost",
			DBPort:                 "5432",
			DBUser:                 "postgres",
			DBPassword:             "somepass",
			DBName:                 "smart_chat",
			Email:                  "test@test.com",
			EmailPassword:          "xxxxxxxxxx",
			SecretToken:            "secret_token",
			IndianTeavellersURL:    "http://127.0.0.1:8000",
			NotificationServiceURL: "http://127.0.0.1:8001",
			SlackNotificationURL:   "https://hooks.slack.com/services/xx",
			SlackAlertURL:          "https://hooks.slack.com/services/xx",
		}
	}

	return config
}
