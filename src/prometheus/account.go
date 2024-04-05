package prometheus

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/big"
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
	// Array to store randomly selected characters
	var usernameBytes []byte
	for i := 0; i < length; i++ {
		// Select a random byte
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}
		// Insert byte into the array
		usernameBytes = append(usernameBytes, alphabet[randomIndex.Int64()])
	}

	return string(usernameBytes)
}
