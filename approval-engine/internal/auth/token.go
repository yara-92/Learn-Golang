// Package auth 提供一个极简的、仅依赖标准库的签名令牌实现（HMAC-SHA256），
// 语义上等价于一个简化版 JWT：payload.signature，两段式，base64url 编码。
//
// 之所以不引入第三方 JWT 库，是为了让整个项目除 SQLite 驱动和 bcrypt 外
// 不再有任何额外依赖，方便离线构建、方便你逐行看懂鉴权到底是怎么回事。
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTokenMalformed = errors.New("token: malformed")
	ErrTokenExpired   = errors.New("token: expired")
	ErrTokenInvalid   = errors.New("token: signature invalid")
)

type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	ExpireAt int64  `json:"exp"` // unix seconds
}

type Signer struct {
	secret []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// Generate 生成一个新的签名令牌，ttl 后过期。
func (s *Signer) Generate(userID int64, username, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		ExpireAt: time.Now().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	sig := s.sign(payloadPart)
	return payloadPart + "." + sig, nil
}

// Parse 校验并解析令牌，返回其中的 Claims。
func (s *Signer) Parse(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrTokenMalformed
	}
	payloadPart, sigPart := parts[0], parts[1]

	expectedSig := s.sign(payloadPart)
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(sigPart)) != 1 {
		return nil, ErrTokenInvalid
	}

	raw, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return nil, ErrTokenMalformed
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrTokenMalformed
	}
	if time.Now().Unix() > claims.ExpireAt {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func (s *Signer) sign(payloadPart string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payloadPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
