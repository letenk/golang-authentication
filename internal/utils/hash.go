package utils

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
    // Generate hash with cost 12
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    
    return string(hashedBytes), nil
}

// ComparePassword compares a plain text password with a hashed password
func ComparePassword(hashedPassword, plainPassword string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}