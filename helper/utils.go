package helper

import (
    "math/rand"
    "time"
)

func RandomString(n int) string {
    letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    rand.Seed(time.Now().UnixNano())

    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func GenerateRandomCode(n int) string {
    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    rand.Seed(time.Now().UnixNano())

    result := make([]byte, n)
    for i := range result {
        result[i] = letters[rand.Intn(len(letters))]
    }
    return string(result)
}