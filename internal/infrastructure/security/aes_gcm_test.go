package security

import (
	"os"
	"testing"
)

func TestAESGCMEncryptor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sec_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	enc, err := NewAESGCMEncryptor(tempDir)
	if err != nil {
		t.Fatalf("error inicializando encryptor: %v", err)
	}

	rawSecret := "sk-ant-api03-very-secret-token-12345"
	cipher, err := enc.Encrypt(rawSecret)
	if err != nil {
		t.Fatalf("error encriptando: %v", err)
	}

	if string(cipher) == rawSecret {
		t.Fatalf("el texto cifrado no debe coincidir con el texto plano")
	}

	decrypted, err := enc.Decrypt(cipher)
	if err != nil {
		t.Fatalf("error descifrando: %v", err)
	}

	if decrypted != rawSecret {
		t.Fatalf("esperaba '%s', obtuve '%s'", rawSecret, decrypted)
	}

	// Probar fallo con bytes corruptos
	if len(cipher) > 15 {
		cipher[15] ^= 0xFF
		_, err = enc.Decrypt(cipher)
		if err == nil {
			t.Fatalf("esperaba error por datos corruptos")
		}
	}
}
