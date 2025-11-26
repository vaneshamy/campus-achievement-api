package helper

import (
    "os"
    "time"
	"fmt"

    "github.com/golang-jwt/jwt/v5"
    "go-fiber/app/model"
)

func GenerateAccessToken(user *model.User, perms []string) (string, error) {
    secret := os.Getenv("JWT_SECRET")

    claims := jwt.MapClaims{
        "userId":      user.ID,
        "username":    user.Username,
        "role":        user.RoleName,
        "permissions": perms,
        "type":        "access",
        "exp":         time.Now().Add(1 * time.Hour).Unix(),
    }

    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(secret))
}

func GenerateRefreshToken(userID string) (string, error) {
    secret := os.Getenv("JWT_SECRET")

    claims := jwt.MapClaims{
        "userId": userID,
        "type":   "refresh",
        "exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
    }

    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(secret))
}

func ValidateRefreshToken(tokenStr string) (string, error) {
    secret := os.Getenv("JWT_SECRET")

    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil || !token.Valid {
        return "", err
    }

    claims := token.Claims.(jwt.MapClaims)
    return claims["userId"].(string), nil
}

func ValidateAccessToken(tokenStr string) (*model.JWTClaims, error) {
    secret := os.Getenv("JWT_SECRET")

    // Parse token
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil || !token.Valid {
        return nil, err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid claims")
    }

    // Validasi bahwa token adalah access token
    if claims["type"] != "access" {
        return nil, fmt.Errorf("invalid token type")
    }

    // Build JWTClaims struct
    permissions := []string{}
    if perms, ok := claims["permissions"].([]interface{}); ok {
        for _, p := range perms {
            if str, ok := p.(string); ok {
                permissions = append(permissions, str)
            }
        }
    }

    return &model.JWTClaims{
        UserID:      claims["userId"].(string),
        Username:    claims["username"].(string),
        Role:        claims["role"].(string),
        Permissions: permissions,
        Type:        "access",
    }, nil
}

