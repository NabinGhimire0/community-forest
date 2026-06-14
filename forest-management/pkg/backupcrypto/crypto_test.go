package backupcrypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.bin")
	encrypted := filepath.Join(dir, "backup.bsmbackup")
	decrypted := filepath.Join(dir, "output.bin")
	content := bytes.Repeat([]byte("Ban Samiti backup test\n"), 100000)
	if err := os.WriteFile(input, content, 0o600); err != nil {
		t.Fatal(err)
	}
	passphrase := "Correct-Horse-Battery-Staple!2026"
	if err := EncryptFile(input, encrypted, passphrase); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if err := DecryptFile(encrypted, decrypted, passphrase); err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	result, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, result) {
		t.Fatal("decrypted content differs from input")
	}
}

func TestDecryptRejectsWrongPassphraseAndRemovesOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	encrypted := filepath.Join(dir, "backup.bsmbackup")
	output := filepath.Join(dir, "output.txt")
	if err := os.WriteFile(input, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(input, encrypted, "A-Strong-Backup-Passphrase!1"); err != nil {
		t.Fatal(err)
	}
	if err := DecryptFile(encrypted, output, "Definitely-Wrong-Passphrase!2"); err == nil {
		t.Fatal("expected wrong passphrase to fail")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("failed decryption must not leave plaintext output")
	}
}

func TestDecryptRejectsTrailingData(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	encrypted := filepath.Join(dir, "backup.bsmbackup")
	output := filepath.Join(dir, "output.txt")
	passphrase := "A-Strong-Backup-Passphrase!1"
	if err := os.WriteFile(input, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(input, encrypted, passphrase); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(encrypted, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("unexpected")); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	_ = file.Close()
	if err := DecryptFile(encrypted, output, passphrase); err == nil {
		t.Fatal("expected trailing data to be rejected")
	}
}
