package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrDecryptionFailed = errors.New("falló el descifrado: datos corruptos o clave incorrecta")
)

type AESGCMEncryptor struct {
	key []byte
}

// NewAESGCMEncryptor inicializa el cifrador derivando una llave de 256 bits a partir de un secret y un salt local
func NewAESGCMEncryptor(dataDir string) (*AESGCMEncryptor, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("error creando directorio de datos: %w", err)
	}

	saltFile := filepath.Join(dataDir, ".salt")
	var salt []byte

	if _, err := os.Stat(saltFile); os.IsNotExist(err) {
		salt = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("error generando salt: %w", err)
		}
		if err := os.WriteFile(saltFile, salt, 0600); err != nil {
			return nil, fmt.Errorf("error guardando salt: %w", err)
		}
	} else {
		var err error
		salt, err = os.ReadFile(saltFile)
		if err != nil {
			return nil, fmt.Errorf("error leyendo salt: %w", err)
		}
	}

	// Obtener machine-id o identificador local
	machineID := "cv-checker-default-host"
	if mid, err := os.ReadFile("/etc/machine-id"); err == nil {
		machineID = string(mid)
	} else if hostname, err := os.Hostname(); err == nil {
		machineID = hostname
	}

	// Derivar llave de 32 bytes (256 bits) usando SHA-256 (KDF simple y determinista)
	h := sha256.New()
	h.Write([]byte(machineID))
	h.Write(salt)
	key := h.Sum(nil)

	return &AESGCMEncryptor{key: key}, nil
}

// Encrypt cifra un texto plano usando AES-256 GCM con un nonce aleatorio
func (e *AESGCMEncryptor) Encrypt(plainText string) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// El resultado contiene nonce + ciphertext autenticado
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return cipherText, nil
}

// Decrypt descifra el ciphertext usando AES-256 GCM
func (e *AESGCMEncryptor) Decrypt(cipherText []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, actualCipher := cipherText[:nonceSize], cipherText[nonceSize:]
	plainBytes, err := gcm.Open(nil, nonce, actualCipher, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plainBytes), nil
}
