package provider

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// TokenCacheFile represents the structure of the cached token file
type TokenCacheFile struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CustomerID string   `json:"customer_id"`
	ClientID  string    `json:"client_id"`
}

// TokenPersistence handles encrypted disk-based token caching
type TokenPersistence struct {
	cacheDir  string
	password  string
}

// NewTokenPersistence creates a new token persistence manager
func NewTokenPersistence(customerID, clientID string) *TokenPersistence {
	// Create cache directory in user's home
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".terraform.d", "spa-cache")
	os.MkdirAll(cacheDir, 0700)
	
	// Use customer ID + client ID as password base for encryption
	password := fmt.Sprintf("%s:%s", customerID, clientID)
	
	return &TokenPersistence{
		cacheDir:  cacheDir,
		password:  password,
	}
}

// getCacheFilePath returns the path to the token cache file for this customer/client combination
func (tp *TokenPersistence) getCacheFilePath(customerID, clientID string) string {
	// Create a safe filename from customer ID and client ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", customerID, clientID)))
	filename := fmt.Sprintf("token_%x.cache", hash[:8])
	return filepath.Join(tp.cacheDir, filename)
}

// encrypt encrypts data using AES-GCM with PBKDF2 key derivation
func (tp *TokenPersistence) encrypt(data []byte) ([]byte, error) {
	// Generate a random salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(tp.password), salt, 100000, 32, sha256.New)
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	
	// Encrypt the data
	encrypted := gcm.Seal(nil, nonce, data, nil)
	
	// Combine salt + nonce + encrypted data
	result := make([]byte, 0, len(salt)+len(nonce)+len(encrypted))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, encrypted...)
	
	return result, nil
}

// decrypt decrypts data using AES-GCM with PBKDF2 key derivation
func (tp *TokenPersistence) decrypt(data []byte) ([]byte, error) {
	if len(data) < 44 { // 32 bytes salt + 12 bytes nonce minimum
		return nil, fmt.Errorf("invalid encrypted data")
	}
	
	// Extract salt, nonce, and encrypted data
	salt := data[:32]
	nonce := data[32:44]
	encrypted := data[44:]
	
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(tp.password), salt, 100000, 32, sha256.New)
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Decrypt the data
	decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	
	return decrypted, nil
}

// LoadToken loads a cached token from disk if it exists and is still valid
func (tp *TokenPersistence) LoadToken(customerID, clientID string) (*TokenCacheFile, error) {
	filePath := tp.getCacheFilePath(customerID, clientID)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil // No cached token
	}
	
	// Read encrypted file
	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token cache file: %w", err)
	}
	
	// Decrypt the data
	decryptedData, err := tp.decrypt(encryptedData)
	if err != nil {
		// If decryption fails, remove the corrupted file
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to decrypt token cache: %w", err)
	}
	
	// Parse JSON
	var tokenCache TokenCacheFile
	if err := json.Unmarshal(decryptedData, &tokenCache); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to parse token cache: %w", err)
	}
	
	// Check if token is still valid (with 5 minute buffer)
	if time.Now().Add(5 * time.Minute).After(tokenCache.ExpiresAt) {
		// Token expired, remove file
		os.Remove(filePath)
		return nil, nil
	}
	
	return &tokenCache, nil
}

// SaveToken saves a token to disk with encryption
func (tp *TokenPersistence) SaveToken(customerID, clientID, token string, expiresAt time.Time) error {
	tokenCache := TokenCacheFile{
		Token:      token,
		ExpiresAt:  expiresAt,
		CustomerID: customerID,
		ClientID:   clientID,
	}
	
	// Marshal to JSON
	jsonData, err := json.Marshal(tokenCache)
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}
	
	// Encrypt the data
	encryptedData, err := tp.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt token cache: %w", err)
	}
	
	// Write to file with secure permissions
	filePath := tp.getCacheFilePath(customerID, clientID)
	if err := os.WriteFile(filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write token cache file: %w", err)
	}
	
	return nil
}

// ClearToken removes the cached token file
func (tp *TokenPersistence) ClearToken(customerID, clientID string) error {
	filePath := tp.getCacheFilePath(customerID, clientID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token cache file: %w", err)
	}
	return nil
}
