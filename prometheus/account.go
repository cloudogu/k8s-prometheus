package prometheus

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/rand/v2"
	"strings"
)

const usernameAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const passwordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"

func createNewAccount(consumer string) (string, string) {
	username := generateRadomString(usernameAlphabet, 8)
	password := generateRadomString(passwordAlphabet, 24)

	return fmt.Sprintf("%s-%s", consumer, username), password
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

func generateRadomString(alphabet string, length int) string {
	runeSlice := []rune(alphabet)
	maxIndex := len(alphabet)

	// Array to store randomly selected characters
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		// Select a random byte
		randomIndex := rand.IntN(maxIndex)
		// Insert byte into the array
		sb.WriteRune(runeSlice[randomIndex])
	}

	return sb.String()
}
