package utils

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int, secret string, expiry time.Duration) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(expiry).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
