package utils

import (
    "encoding/base64"
    "os"
)

func LoadKey() []byte {
    keyB64 := os.Getenv("FILE_ENC_KEY")
    if keyB64 == "" {
        Error.Debug().Msg("FILE_ENC_KEY not set")
    }

    key, err := base64.StdEncoding.DecodeString(keyB64)
    if err != nil {
        Error.Fatal().Str("invalid base64 key:", err.Error())
    }

    if len(key) != 32 {
        Error.Fatal().Msg("key must be exactly 32 bytes")
    }

    return key
}
