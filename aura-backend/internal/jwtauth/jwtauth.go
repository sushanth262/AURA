package jwtauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

type Claims struct {
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

type Principal struct {
	Sub      string
	Roles    []string
	TenantID string
}

func VerifyBearer(ctx context.Context, cfg config.Config, bearer string) (Principal, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(bearer), "Bearer ")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Principal{}, errors.New("missing bearer token")
	}

	if cfg.AuthDevMock {
		return verifyHS256(cfg, raw)
	}

	if cfg.AuthJWKSURL != "" {
		return Principal{}, errors.New("OIDC/JWKS mode not wired in this build — set AUTH_DEV_MOCK=true or extend jwtauth with JWKS")
	}

	return Principal{}, fmt.Errorf("AUTH_DEV_MOCK=false requires AUTH_JWKS_URL (JWKS validation not implemented in stub)")
}

func verifyHS256(cfg config.Config, raw string) (Principal, error) {
	if len(cfg.AuthDevJWTSecret) < 16 {
		return Principal{}, errors.New("AUTH_DEV_JWT_SECRET must be at least 16 characters")
	}
	aud := cfg.AuthAudience
	if aud == "" {
		aud = "aura-api"
	}
	iss := cfg.AuthIssuer
	if iss == "" {
		iss = "aura-dev"
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(aud),
		jwt.WithIssuer(iss),
		jwt.WithLeeway(30*time.Second),
	)
	token, err := parser.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(cfg.AuthDevJWTSecret), nil
	})
	if err != nil {
		return Principal{}, err
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Principal{}, errors.New("invalid token claims")
	}
	sub := c.Subject
	if sub == "" {
		return Principal{}, errors.New("token missing sub")
	}
	return Principal{
		Sub:      sub,
		Roles:    c.Roles,
		TenantID: c.TenantID,
	}, nil
}

// MintDevToken issues a short-lived HS256 JWT for local/demo use only.
func MintDevToken(cfg config.Config, sub string, roles []string, tenantID string) (string, error) {
	if len(cfg.AuthDevJWTSecret) < 16 {
		return "", errors.New("AUTH_DEV_JWT_SECRET must be at least 16 characters")
	}
	now := time.Now()
	c := Claims{
		Roles:    roles,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
		},
	}
	if cfg.AuthIssuer == "" {
		c.RegisteredClaims.Issuer = "aura-dev"
	}
	if cfg.AuthAudience == "" {
		c.RegisteredClaims.Audience = jwt.ClaimStrings{"aura-api"}
	} else {
		c.RegisteredClaims.Audience = jwt.ClaimStrings{cfg.AuthAudience}
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	return t.SignedString([]byte(cfg.AuthDevJWTSecret))
}
