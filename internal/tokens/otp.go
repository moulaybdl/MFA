package tokens

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	OTPToken = "otp"
)

var (
	ErrOTPNotMatch = errors.New("the OTP does not match")
)

func GenerateOTP() (string, error) {
	randomByte := make([]byte, 16)

	_, err := rand.Read(randomByte)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte), nil
}

func SetOTPCache(otp string, userID int, client *redis.Client, r *http.Request) error {
	field := fmt.Sprintf("%d", userID)

	err := client.HSet(r.Context(), OTPToken, map[string]string{
		field: otp,
	}, time.Minute * 5).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetOTPCache(userID int, client *redis.Client, r *http.Request) (string, error) {
	field := fmt.Sprintf("%d", userID)

	val, err := client.HGet(r.Context(), OTPToken, field).Result()
	if err == nil {
		return "", err
	}

	return val, nil

}

func VerifyOTPMatch(otpCache, otpUser string) error {
	if otpCache != otpUser {
		return ErrOTPNotMatch
	}
	return nil
}