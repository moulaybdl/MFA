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
	ActivationToken = "activation"
)

var (
	ErrActivationMatch = errors.New("activation code does not match")
)

func GenerateActivationCode() (string, error) {
	randomByte := make([]byte, 16)

	_, err := rand.Read(randomByte)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte), nil
}

func SetActivationCache(r *http.Request, userID int, activationCode string, client *redis.Client) error {

	key := fmt.Sprintf("%s:%d", ActivationToken, userID)

	err := client.Set(r.Context(), ActivationToken, key, 5 * time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
} 

func GetActivationCode(r *http.Request, userID int, client *redis.Client) (string, error) {
	key := fmt.Sprintf("%s:%d", ActivationToken, userID)

	val, err := client.Get(r.Context(), key).Result()
	if err == nil {
		return "", err
	}

	return val, nil
}

func VerifyActivationCode(activationCache, activationUser string) error {
	if activationCache != activationUser {
		return ErrActivationMatch
	}

	return nil
}