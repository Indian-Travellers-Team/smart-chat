package zitadel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, rawToken string) (*ValidateTokenUser, error)
}

type Service struct {
	client           *resty.Client
	validateTokenURL string
}

func NewService(_ context.Context, cfg ZitadelConfig) (*Service, error) {
	baseURL := strings.TrimSpace(cfg.AuthServiceBaseURL)
	if baseURL == "" {
		return nil, errors.New("auth service base url is required")
	}

	validatePath := strings.TrimSpace(cfg.ValidateTokenPath)
	if validatePath == "" {
		validatePath = "/api/v1/token/validate"
	}
	if !strings.HasPrefix(validatePath, "/") {
		validatePath = "/" + validatePath
	}

	return &Service{
		client: resty.New().
			SetBaseURL(strings.TrimRight(baseURL, "/")).
			SetTimeout(10 * time.Second),
		validateTokenURL: validatePath,
	}, nil

}

func (s *Service) ValidateToken(ctx context.Context, rawToken string) (*ValidateTokenUser, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, errors.New("token is required")
	}

	requestPayload := ValidateTokenRequest{Token: &rawToken}
	var responsePayload ValidateTokenResponse

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(requestPayload).
		SetResult(&responsePayload).
		Post(s.validateTokenURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service validate api: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("auth service validate api returned status %d", resp.StatusCode())
	}

	if responsePayload.Error != nil && strings.TrimSpace(*responsePayload.Error) != "" {
		return nil, errors.New(strings.TrimSpace(*responsePayload.Error))
	}

	if responsePayload.User == nil {
		return nil, errors.New("invalid auth service response")
	}

	return responsePayload.User, nil
}
