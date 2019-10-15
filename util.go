package ispend

import (
	"encoding/json"
	"io"
	"math"
	"math/rand"

	"golang.org/x/crypto/bcrypt"
)

func GenerateRandomString(length int) string {
	text := ""
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	for i := 0; i < length; i++ {
		possibleLen := float64(len(possible))
		nextPossible := math.Floor(rand.Float64() * possibleLen)
		text += string(possible[int(nextPossible)])
	}

	return text
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SendAPIResp(w io.Writer, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func SendAPIOKResp(w io.Writer, message string) error {
	apiResp := APIResponse{Status: 200, Message: message}
	return SendAPIResp(w, apiResp)
}

func SendAPIOKRespWithData(w io.Writer, message string, data interface{}) error {
	apiResp := APIResponse{Status: 200, Message: message, Data: data}
	return SendAPIResp(w, apiResp)
}

func SendAPIErrorResp(w io.Writer, message string, status int) error {
	apiErr := APIResponse{Status: status, Message: message, IsError: true}
	return SendAPIResp(w, apiErr)
}
