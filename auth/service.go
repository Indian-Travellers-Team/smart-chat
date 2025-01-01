package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"smart-chat/config"
	"smart-chat/internal/store" // Update with the correct package path
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type AuthService struct {
	Fast2SMSAPIKey string
}

func NewAuthService() *AuthService {
	cfg := config.Load()
	return &AuthService{
		Fast2SMSAPIKey: cfg.FAST2SMS_API_KEY,
	}
}

func init() {
	cfg := config.Load()
	fast2SMSAPIKey := cfg.FAST2SMS_API_KEY
	if fast2SMSAPIKey == "" {
		log.Fatal("fast2SMSAPIKey is not set in environment variables")
	}
}

func (s *AuthService) InitLogin(info UserLoginInfo) (gin.H, error) {
	// Generate OTP
	otp, err := generateOTP(4) // 4-digit OTP
	if err != nil {
		return gin.H{"error": err, "success": false}, err
	}

	log.Printf("Sending OTP '%s' to mobile: %s", otp, info.Mobile)
	otp_err := s.sendOTPMessage(info.Mobile, otp)
	if otp_err != nil {
		return gin.H{"error": "failed to send OTP", "success": false}, otp_err
	}

	authToken, err := generateSecretToken(32) // 32 bytes before base64 encoding
	if err != nil {
		return gin.H{"error": "failed to generate auth token", "success": false}, err
	}

	// Store authToken and OTP in the repository
	user := store.User{
		Name:           info.Name,
		Mobile:         info.Mobile,
		OTP:            otp,
		AuthToken:      authToken,
		AuthExpireTime: time.Now().Add(5 * time.Minute), // Example expiry time
	}
	err = store.StoreUser(user)
	if err != nil {
		return gin.H{"error": err, "success": false}, err
	}

	return gin.H{"authToken": authToken, "success": true}, nil
}

func (s *AuthService) ValidateLogin(authTokenString string, otp string) (gin.H, error) {
	// Retrieve the user by authToken
	user, err := store.GetUser(authTokenString)
	if err != nil {
		return gin.H{"error": err}, errors.New("invalid auth token")
	}

	// Check if OTP matches and auth token is not expired
	if user.OTP != otp || time.Now().After(user.AuthExpireTime) {
		return gin.H{"error": err, "success": false}, errors.New("invalid OTP or token expired")
	}

	// Generate an access token (this should be unique and securely generated)
	accessTokenString, err := generateSecretToken(32) // 32 bytes before base64 encoding
	if err != nil {
		return gin.H{"error": "failed to generate access token", "success": false}, err
	}
	// Create and store a new conversation with the generated access token
	conversation := store.Conversation{
		AccessToken:           accessTokenString,
		AccessTokenExpireTime: time.Now().Add(1 * time.Hour),
		UserName:              user.Name,
		UserMobile:            user.Mobile,
		Messages:              []store.MessagePair{},
		Pushed:                false,
		TotalTokens:           0,
	}
	err = store.StoreConversation(accessTokenString, conversation)
	if err != nil {
		return gin.H{"error": err, "success": false}, err
	}

	return gin.H{"accessToken": accessTokenString, "success": true}, nil
}

func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)

	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}

func (s *AuthService) sendOTPMessage(mobile, otp string) error {
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

func generateSecretToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
