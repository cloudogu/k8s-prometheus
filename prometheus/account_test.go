package prometheus

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
)

func Test_createNewAccount(t *testing.T) {
	t.Run("should create new account for consumer", func(t *testing.T) {
		username, password, err := createNewAccount("myConsumer")
		require.NoError(t, err)

		assert.Len(t, username, 19)
		assert.True(t, strings.HasPrefix(username, "myConsumer-"))
		assert.Len(t, password, 24)

		username2, password2, err := createNewAccount("myConsumer")
		require.NoError(t, err)

		assert.NotEqual(t, username, username2)
		assert.Len(t, username2, 19)
		assert.True(t, strings.HasPrefix(username2, "myConsumer-"))
		assert.NotEqual(t, password, password2)
		assert.Len(t, password2, 24)
	})
}

func Test_HashPassword(t *testing.T) {
	t.Run("should hash password", func(t *testing.T) {
		hashedPassword, err := hashPassword("password123")

		require.NoError(t, err)
		assert.NotEqual(t, "", hashedPassword)
		assert.Nil(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("password123")))
	})

	t.Run("should fail for passwords longer than 72 bytes", func(t *testing.T) {
		_, err := hashPassword("password123password123password123password123password123password123password123")

		require.Error(t, err)
		assert.ErrorContains(t, err, "error generating hash of password: bcrypt: password length exceeds 72 bytes")
	})
}

func Test_CompareHashAndPassword(t *testing.T) {
	t.Run("should succeed for correct password", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("otherPassword123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		err = compareHashAndPassword(string(hashedPassword), "otherPassword123")

		require.NoError(t, err)
	})

	t.Run("should fail for incorrect password", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("otherPassword123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		err = compareHashAndPassword(string(hashedPassword), "foBar")

		require.Error(t, err)
		assert.ErrorContains(t, err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	})
}

func Test_GenerateRandomString(t *testing.T) {
	t.Run("should generate string for alphabet", func(t *testing.T) {
		randomString, err := generateRadomString("abc", 5)
		require.NoError(t, err)

		require.Len(t, randomString, 5)
		assert.True(t, strings.ContainsAny(randomString, "abc"))
		for _, char := range randomString {
			if char != 'a' && char != 'b' && char != 'c' {
				t.Errorf("the generated string contains a not-allowed character: %c", char)
			}
		}
	})
}
