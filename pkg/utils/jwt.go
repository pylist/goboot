package utils

import (
	"errors"
	"goboot/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    uint      `json:"userId"`
	Username  string    `json:"username"`
	Role      int8      `json:"role"`
	TokenType TokenType `json:"tokenType"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // Access Token过期时间(秒)
}

// GenerateTokenPair 生成双Token
func GenerateTokenPair(userID uint, username string, role int8) (*TokenPair, error) {
	accessToken, err := generateToken(userID, username, role, AccessToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(userID, username, role, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire) * 3600,
	}, nil
}

func generateToken(userID uint, username string, role int8, tokenType TokenType) (string, error) {
	cfg := config.AppConfig.JWT

	var expire int
	var secret string

	if tokenType == AccessToken {
		expire = cfg.AccessExpire
		secret = cfg.Secret
	} else {
		expire = cfg.RefreshExpire
		secret = cfg.RefreshSecret
	}

	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAccessToken 解析Access Token
func ParseAccessToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.AppConfig.JWT.Secret, AccessToken)
}

// ParseRefreshToken 解析Refresh Token
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.AppConfig.JWT.RefreshSecret, RefreshToken)
}

func parseToken(tokenString, secret string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.TokenType != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// RefreshAccessToken 使用Refresh Token刷新Access Token
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := ParseRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
}

// 兼容旧接口
func GenerateToken(userID uint, username string, role int8) (string, error) {
	return generateToken(userID, username, role, AccessToken)
}

func ParseToken(tokenString string) (*Claims, error) {
	return ParseAccessToken(tokenString)
}
