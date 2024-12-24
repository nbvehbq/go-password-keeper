package client

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/nbvehbq/go-password-keeper/internal/logger"
	"github.com/nbvehbq/go-password-keeper/internal/model"
	"go.uber.org/zap"
)

var (
	ErrUnauthorized = fmt.Errorf("unauthorized")
	ErrInternal     = fmt.Errorf("internal error")
	ErrUserExists   = fmt.Errorf("user already exists")
	ErrGenerateKey  = fmt.Errorf("failed to generate key")
	ErrLoadKeyFile  = fmt.Errorf("failed to load key file")
	ErrEncrypt      = fmt.Errorf("failed to encrypt data")
	ErrDecrypt      = fmt.Errorf("failed to decrypt data")
)

type Client struct {
	client     *resty.Client
	cfg        *Config
	privateKey []byte
}

func NewClient(ctx context.Context, config *Config) (*Client, error) {
	client := resty.New()
	// client.SetDebug(true)
	client.SetHeader("Accept", "application/json")

	return &Client{
		client: client,
		cfg:    config,
	}, nil
}

func (c *Client) Register(ctx context.Context, login, password string) error {
	var payload struct {
		SID string `json:"sid"`
	}
	res, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"login": login, "password": password}).
		Post(fmt.Sprintf("%s/api/user/register", c.cfg.Address))

	if err != nil {
		logger.Log.Error("failed to register", zap.Error(err))
		return ErrInternal
	}

	if res.StatusCode() == 409 {
		return ErrUserExists
	}

	// Set credentials
	c.client.SetHeader("Authorization", payload.SID)
	c.client.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    payload.SID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Generate sertificate
	cert, err := setupKeyPair()
	if err != nil {
		logger.Log.Error("failed to generate key", zap.Error(err))
		return ErrGenerateKey
	}

	c.privateKey = cert

	// check directory & create if not exists
	if _, err := os.Stat(c.cfg.KeyPath); os.IsNotExist(err) {
		if err := os.Mkdir(c.cfg.KeyPath, 0777); err != nil {
			logger.Log.Error("failed to create directory", zap.Error(err))
			return ErrGenerateKey
		}
	}

	f, err := os.OpenFile(fmt.Sprintf("%s%s-cert.pem", c.cfg.KeyPath, login), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Log.Error("failed to create key file", zap.Error(err))
		return ErrGenerateKey
	}
	defer f.Close()

	// write cert
	if _, err := f.Write(cert); err != nil {
		logger.Log.Error("failed to write key file", zap.Error(err))
		return ErrGenerateKey
	}

	return nil
}

func (c *Client) Login(ctx context.Context, login, password string) error {
	var payload struct {
		SID string `json:"sid"`
	}
	res, err := c.client.R().
		SetContext(ctx).
		SetResult(&payload).
		SetBody(map[string]string{"login": login, "password": password}).
		Post(fmt.Sprintf("%s/api/user/login", c.cfg.Address))

	if err != nil {
		logger.Log.Error("failed to login", zap.Error(err))
		return ErrInternal
	}

	if res.StatusCode() == 401 {
		return ErrUnauthorized
	}

	// Set credentials
	c.client.SetHeader("Authorization", payload.SID)
	c.client.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    payload.SID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Load sertificate
	cert, err := os.ReadFile(fmt.Sprintf("%s%s-cert.pem", c.cfg.KeyPath, login))
	if err != nil {
		logger.Log.Error("failed to read key file", zap.Error(err))
		return ErrLoadKeyFile
	}

	c.privateKey = cert

	return nil
}

func (c *Client) ListSecrets(ctx context.Context, resourceType string) ([]model.Secret, error) {
	var result []model.Secret
	res, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("type", resourceType).
		SetResult(&result).
		Get(fmt.Sprintf("%s/api/secret", c.cfg.Address))

	if err != nil {
		logger.Log.Error("failed to list secrets", zap.Error(err))
		return nil, err
	}

	if res.StatusCode() == 401 {
		return nil, ErrUnauthorized
	}

	return result, nil
}

func (c *Client) CreateSecret(ctx context.Context, data *model.Secret) (int64, error) {
	var response struct {
		ID int64 `json:"id"`
	}

	// encrypt payload & meta
	var err error
	data.Payload, err = encrypt(c.privateKey, data.Payload)
	if err != nil {
		logger.Log.Error("failed to encrypt data", zap.Error(err))
		return 0, ErrEncrypt
	}
	data.Meta, err = encrypt(c.privateKey, data.Meta)
	if err != nil {
		logger.Log.Error("failed to encrypt data", zap.Error(err))
		return 0, ErrEncrypt
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		SetBody(data).
		Post(fmt.Sprintf("%s/api/secret", c.cfg.Address))

	if err != nil {
		logger.Log.Error("failed to create secret", zap.Error(err))
		return 0, err
	}

	if res.StatusCode() == 401 {
		return 0, ErrUnauthorized
	}

	return response.ID, nil
}

func (c *Client) GetSecret(ctx context.Context, ID int64) (*model.Secret, error) {
	var secret model.Secret
	res, err := c.client.R().
		SetContext(ctx).
		SetResult(&secret).
		Get(fmt.Sprintf("%s/api/secret/%d", c.cfg.Address, ID))

	if err != nil {
		logger.Log.Error("failed to get secret", zap.Error(err))
		return nil, err
	}

	if res.StatusCode() == 401 {
		return nil, ErrUnauthorized
	}

	// decrypt payload & meta
	secret.Payload, err = decrypt(c.privateKey, secret.Payload)
	if err != nil {
		logger.Log.Error("failed to decrypt data", zap.Error(err))
		return nil, ErrDecrypt
	}
	secret.Meta, err = decrypt(c.privateKey, secret.Meta)
	if err != nil {
		logger.Log.Error("failed to decrypt data", zap.Error(err))
		return nil, ErrDecrypt
	}

	return &secret, nil
}

func (c *Client) DeleteSecret(ctx context.Context, ID int64) error {
	res, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("%s/api/secret/%d", c.cfg.Address, ID))

	if err != nil {
		logger.Log.Error("failed to delete secret", zap.Error(err))
		return err
	}

	if res.StatusCode() == 401 {
		return ErrUnauthorized
	}

	return nil
}

func (c *Client) UpdateSecret(ctx context.Context, id int64, data *model.Secret) (int64, error) {
	var newID int64
	res, err := c.client.R().
		SetContext(ctx).
		SetResult(&newID).
		SetBody(data).
		Put(fmt.Sprintf("%s/api/secret/%d", c.cfg.Address, id))

	if err != nil {
		logger.Log.Error("failed to update secret", zap.Error(err))
		return 0, err
	}

	if res.StatusCode() == 401 {
		return 0, ErrUnauthorized
	}

	// decrypt payload & meta
	data.Payload, err = decrypt(c.privateKey, data.Payload)
	if err != nil {
		logger.Log.Error("failed to decrypt data", zap.Error(err))
		return 0, ErrDecrypt
	}
	data.Meta, err = decrypt(c.privateKey, data.Meta)
	if err != nil {
		logger.Log.Error("failed to decrypt data", zap.Error(err))
		return 0, ErrDecrypt
	}

	return newID, nil
}
