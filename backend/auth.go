package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type DeviceInfo struct {
	SourceDeviceId string `json:"sourceDeviceId"`
	SourceType     string `json:"sourceType"`
	AppVersion     string `json:"appVersion"`
	MetaDetails    struct {
		UserAgent string `json:"userAgent"`
	} `json:"metaDetails"`
}

type RefreshTokenRequest struct {
	RefreshToken string     `json:"refreshToken"`
	DeviceInfo   DeviceInfo `json:"deviceInfo"`
}

type TokenResponse struct {
	RefreshToken  string `json:"refreshToken"`
	Token         string `json:"token"`
	TokenExpireIn string `json:"tokenExpireIn"`
}

var (
	token           string
	refreshToken    string
	tokenExpireIn   time.Time
	debugMode       bool
	fnsApiUrl       string
	fnsDeviceID     string
	userAgentString string
)

func init() {
	refreshToken = os.Getenv("REFRESH_TOKEN")
	debugMode = os.Getenv("DEBUG_MODE") == "true"
	fnsApiUrl = os.Getenv("FNS_API_URL")
	fnsDeviceID = os.Getenv("FNS_DEVICE_ID")
	userAgentString = os.Getenv("USER_AGENT")
}

func refreshAccessToken() error {
	deviceInfo := DeviceInfo{
		SourceDeviceId: fnsDeviceID,
		SourceType:     "WEB",
		AppVersion:     "1.0.0",
		MetaDetails: struct {
			UserAgent string `json:"userAgent"`
		}{
			UserAgent: userAgentString,
		},
	}
	reqBody := RefreshTokenRequest{
		RefreshToken: refreshToken,
		DeviceInfo:   deviceInfo,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fnsApiUrl+"/api/v1/auth/token", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if debugMode {
		log.Printf("DEBUG: Request Type: POST")
		log.Printf("DEBUG: Request URL: %s", req.URL)
		log.Printf("DEBUG: Request Body: %s", string(reqBodyBytes))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if debugMode {
		log.Printf("DEBUG: Response Status: %s", resp.Status)
		log.Printf("DEBUG: Response Body: %s", string(body))
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return err
	}

	token = tokenResponse.Token
	refreshToken = tokenResponse.RefreshToken

	if tokenResponse.TokenExpireIn == "" {
		tokenExpireIn = time.Now().Add(1 * time.Hour)
	} else {
		tokenExpireIn, err = time.Parse(time.RFC3339, tokenResponse.TokenExpireIn)
		if err != nil {
			return err
		}
	}

	os.Setenv("REFRESH_TOKEN", refreshToken)
	os.Setenv("TOKEN_EXPIRE_IN", tokenExpireIn.Format(time.RFC3339))

	if debugMode {
		log.Printf("DEBUG: Successfully obtained access token: %s", token)
		log.Printf("DEBUG: Access token expires at: %s", tokenExpireIn.Format(time.RFC3339))
	}

	return nil
}

func checkTokenExpiration() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" || time.Now().After(tokenExpireIn) {
			err := refreshAccessToken()
			if err != nil {
				c.String(http.StatusInternalServerError, "Error refreshing token: %v", err)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
