package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

// JWTClaims 自定义 JWT 载荷
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// IssueJWT 为用户签发 JWT token (默认有效期 24h)
func IssueJWT(userID int, username string, role int) (string, error) {
	return IssueJWTWithExpiry(userID, username, role, 24*time.Hour)
}

// IssueJWTWithExpiry 签发自定义过期时间的 JWT token
func IssueJWTWithExpiry(userID int, username string, role int, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    config.SystemName,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// VerifyJWT 验证 JWT token 并返回用户信息（含用户状态检查）
// 返回 nil 表示 token 无效、过期、或用户已被禁用/封禁
func VerifyJWT(tokenString string) (*model.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt validation failed: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}

	// 查询数据库验证用户当前状态（用户可能在 JWT 签发后被禁用）
	user, err := model.GetUserById(claims.UserID, false)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user.Status == model.UserStatusDisabled {
		return nil, fmt.Errorf("user is disabled")
	}

	return user, nil
}

// VerifyJWTClaims 仅验证 JWT 签名和过期，不查询数据库
// 适用于高频调用场景——调用方自行处理用户状态检查
func VerifyJWTClaims(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt validation failed: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}

	return claims, nil
}
