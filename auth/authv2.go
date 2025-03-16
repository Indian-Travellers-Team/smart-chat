package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"smart-chat/config"
	"smart-chat/internal/constants"
	"smart-chat/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

type UserLoginInfoWA struct {
	Name        string `json:"name" binding:"required"`
	Mobile      string `json:"mobile" binding:"required"`
	SecretToken string `json:"secret_token" binding:"required"`
}

type RefreshTokenInfo struct {
	Mobile      string `json:"mobile" binding:"required"`
	SecretToken string `json:"secret_token" binding:"required"`
}

type AuthV2Service struct {
	DB             *gorm.DB
	Fast2SMSAPIKey string
	SecretToken    string
}

func NewAuthV2Service(db *gorm.DB) *AuthV2Service {
	cfg := config.Load()
	return &AuthV2Service{
		DB:             db,
		Fast2SMSAPIKey: cfg.FAST2SMS_API_KEY,
		SecretToken:    cfg.SecretToken,
	}
}

func (s *AuthV2Service) NewLoginWA(info UserLoginInfoWA) (gin.H, error) {
	// Check if info.SecretToken matches the configured SecretToken
	if info.SecretToken != s.SecretToken {
		return gin.H{"error": "Invalid secret token", "success": false}, errors.New("invalid secret token")
	}

	// Check if the user already exists by mobile number
	var user models.User
	result := s.DB.Where("mobile = ?", info.Mobile).First(&user)

	// If the user doesn't exist, create a new one
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Creating a new user with all required fields
		user = models.User{
			Name:           info.Name,
			Mobile:         info.Mobile,
			OTP:            "",                            // Empty OTP as we are bypassing OTP logic
			AccessToken:    "",                            // Access token will be generated later
			AccessExpireAt: time.Now().Add(1 * time.Hour), // Set access expiration time (adjust as needed)
		}

		// Save the new user in the database
		if err := s.DB.Create(&user).Error; err != nil {
			return gin.H{"error": err.Error(), "success": false}, err
		}
	} else {
		// If the user exists, update the existing information (optional)
		user.Name = info.Name
		// Save or update user details
		if err := s.DB.Save(&user).Error; err != nil {
			return gin.H{"error": err.Error(), "success": false}, err
		}
	}

	// Generate an access token for the user
	accessToken, err := generateSecretToken(32)
	if err != nil {
		return gin.H{"error": "Failed to generate access token", "success": false}, err
	}

	// Update user with the generated access token and set its expiration time
	user.AccessToken = accessToken
	user.AccessExpireAt = time.Now().Add(7 * 24 * time.Hour)

	// Save the user with the updated access token
	if err := s.DB.Save(&user).Error; err != nil {
		return gin.H{"error": "Failed to update user with access token", "success": false}, err
	}

	// Create a session for the user
	session := models.Session{
		UserID:    user.ID,
		AuthToken: accessToken,
		Source:    constants.WhatsAppSource,
		ExpireAt:  time.Now().Add(1 * time.Hour),
	}

	if err := s.DB.Create(&session).Error; err != nil {
		return gin.H{"error": "Failed to create session", "success": false}, err
	}

	// Return the access token
	return gin.H{"accessToken": accessToken, "success": true}, nil
}

func (s *AuthV2Service) InitLogin(info UserLoginInfo) (gin.H, error) {
	otp, err := generateOTP(4) // 4-digit OTP
	if err != nil {
		return gin.H{"error": err.Error(), "success": false}, err
	}

	// Assuming the mobile number is unique for each user
	var user models.User
	result := s.DB.Where("mobile = ?", info.Mobile).First(&user)
	accessToken, err := generateSecretToken(32)
	if err != nil {
		return gin.H{"error": "Failed to generate access token"}, err
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create a new user if not exists
		user = models.User{
			Name:           info.Name,
			Mobile:         info.Mobile,
			OTP:            otp,
			AccessToken:    accessToken,
			AccessExpireAt: time.Now().Add(5 * time.Minute), // Example expiry time
		}
	} else {
		// Update existing user's OTP and expiry
		user.OTP = otp
		user.AccessToken = accessToken
		user.AccessExpireAt = time.Now().Add(5 * time.Minute)
	}
	// Save or update user in the database
	if err := s.DB.Save(&user).Error; err != nil {
		return gin.H{"error": err.Error(), "success": false}, err
	}

	log.Printf("otp: %v", otp)

	otp_err := s.sendOTPMessage(info.Mobile, otp)
	if otp_err != nil {
		return gin.H{"error": "failed to send OTP", "success": false}, otp_err
	}

	return gin.H{"authToken": accessToken, "success": true}, nil
}

func (s *AuthV2Service) ValidateLogin(accessToken, otp string) (gin.H, error) {
	var user models.User
	result := s.DB.Where("access_token = ? AND otp = ? AND access_expire_at > ?", accessToken, otp, time.Now()).First(&user)
	if result.Error != nil {
		return gin.H{"error": "Invalid OTP or expired"}, errors.New("invalid OTP or expired")
	}

	// Generate a new session for the user
	authToken, err := generateSecretToken(32)
	if err != nil {
		return gin.H{"error": "Failed to generate session token"}, err
	}

	session := models.Session{
		UserID:    user.ID,
		AuthToken: authToken,
		Source:    constants.WebsiteSource,
		ExpireAt:  time.Now().Add(1 * time.Hour),
	}

	log.Printf("session %v", session)

	if err := s.DB.Create(&session).Error; err != nil {
		return gin.H{"error": "Failed to create session"}, err
	}

	return gin.H{"accessToken": authToken, "success": true}, nil
}

func (s *AuthV2Service) sendOTPMessage(mobile, otp string) error {
	client := resty.New()
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"authorization":    s.Fast2SMSAPIKey,
			"variables_values": otp,
			"route":            "otp",
			"numbers":          mobile[3:],
		}).
		SetHeaders(map[string]string{
			"cache-control": "no-cache",
		}).
		Get("https://www.fast2sms.com/dev/bulkV2")

	if err != nil {
		log.Printf("Error sending OTP message: %v", err)
		return err
	}

	// Log response body for debugging
	respString := resp.String()
	log.Printf("Response from Fast2SMS: %s", respString)

	// Parse the response JSON
	var responseJSON map[string]interface{}
	if err := json.Unmarshal([]byte(respString), &responseJSON); err != nil {
		log.Printf("Error parsing response JSON: %v", err)
		return err
	}

	// Check if the "return" key is present and false
	if val, ok := responseJSON["return"].(bool); ok && !val {
		log.Printf("Failed to send OTP: return is false")
		return errors.New("failed to send OTP")
	}

	// Check if response status code is not 200
	if resp.StatusCode() != http.StatusOK {
		log.Printf("Failed to send OTP: Status Code %d", resp.StatusCode())
		return errors.New("failed to send OTP")
	}

	return nil
}

func (s *AuthV2Service) RefreshToken(info RefreshTokenInfo, accessToken string) (gin.H, error) {
	if info.SecretToken != s.SecretToken {
		return gin.H{"error": "Invalid secret token", "success": false}, errors.New("invalid secret token")
	}
	var session models.Session
	result := s.DB.Where("auth_token = ?", accessToken).First(&session)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Println("Session not found for token:", accessToken)
			return gin.H{"error": "session not found"}, result.Error
		} else {
			log.Println("Database error:", result.Error)
			return gin.H{"error": "internal server error"}, result.Error
		}
	}
	if time.Now().After(session.ExpireAt) {
		log.Println("Session expired for user:", session.User.ID)
		authToken, err := generateSecretToken(32)
		if err != nil {
			return gin.H{"error": "Failed to generate session token"}, err
		}
		session.AuthToken = authToken
		session.ExpireAt = time.Now().Add(7 * 24 * time.Hour)
		if err := s.DB.Save(&session).Error; err != nil {
			return gin.H{"error": "Failed to update session"}, err
		}
		return gin.H{"accessToken": authToken, "success": true}, nil
	}
	return gin.H{"error": "session not expired"}, nil
}
