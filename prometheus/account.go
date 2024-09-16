package prometheus

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"strings"
)

const usernameAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const passwordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"

func createNewAccount(consumer string) (string, string, error) {
	username, err := generateRadomString(usernameAlphabet, 8)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate username: %w", err)
	}
	password, err := generateRadomString(passwordAlphabet, 24)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate password: %w", err)
	}

	return fmt.Sprintf("%s-%s", consumer, username), password, nil
}

func hashPassword(password string) (string, error) {
	// Generate a bcrypt hash of the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error generating hash of password: %w", err)
	}

	return string(hashedPassword), nil
}

func compareHashAndPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func generateRadomString(alphabet string, length int) (string, error) {
	runeSlice := []rune(alphabet)
	maxIndex := len(alphabet)

	// Array to store randomly selected characters
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		// Select a random byte
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(maxIndex)))
		if err != nil {
			return "", fmt.Errorf("failed to create random integer: %w", err)
		}
		// Insert byte into the array
		sb.WriteRune(runeSlice[randomIndex.Int64()])
	}

	return sb.String(), nil
}
