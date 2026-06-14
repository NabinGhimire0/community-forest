package backupcrypto

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/scrypt"
)

var magic = []byte("BSMBACKUP1\n")

const (
	saltSize  = 16
	chunkSize = 1024 * 1024
	scryptN   = 32768
	scryptR   = 8
	scryptP   = 1
)

func EncryptFile(inputPath, outputPath, passphrase string) error {
	if len(passphrase) < 16 {
		return errors.New("backup passphrase must contain at least 16 characters")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			_ = os.Remove(outputPath)
		}
	}()

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	key, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, 32)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	writer := bufio.NewWriterSize(output, 256*1024)
	if _, err := writer.Write(magic); err != nil {
		return err
	}
	if _, err := writer.Write(salt); err != nil {
		return err
	}
	for _, value := range []uint32{scryptN, scryptR, scryptP, chunkSize} {
		if err := binary.Write(writer, binary.BigEndian, value); err != nil {
			return err
		}
	}

	buffer := make([]byte, chunkSize)
	var index uint64
	for {
		n, readErr := io.ReadFull(input, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return readErr
		}
		if n > 0 {
			nonce := make([]byte, gcm.NonceSize())
			if _, err := rand.Read(nonce); err != nil {
				return err
			}
			aad := recordAAD(index, uint32(n), false)
			ciphertext := gcm.Seal(nil, nonce, buffer[:n], aad)
			if err := writeRecord(writer, index, uint32(n), nonce, ciphertext); err != nil {
				return err
			}
			index++
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	finalCiphertext := gcm.Seal(nil, nonce, nil, recordAAD(index, 0, true))
	if err := writeRecord(writer, index, 0, nonce, finalCiphertext); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := output.Sync(); err != nil {
		return err
	}
	ok = true
	return nil
}

func DecryptFile(inputPath, outputPath, passphrase string) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()
	reader := bufio.NewReaderSize(input, 256*1024)

	readMagic := make([]byte, len(magic))
	if _, err := io.ReadFull(reader, readMagic); err != nil || string(readMagic) != string(magic) {
		return errors.New("not a Ban Samiti encrypted backup")
	}
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return err
	}
	var n, r, p, configuredChunk uint32
	for _, target := range []*uint32{&n, &r, &p, &configuredChunk} {
		if err := binary.Read(reader, binary.BigEndian, target); err != nil {
			return err
		}
	}
	// Bound all attacker-controlled KDF and allocation parameters before doing
	// expensive work. N must be a power of two as required by scrypt.
	if n < 16384 || n > 1048576 || n&(n-1) != 0 || r == 0 || r > 32 || p == 0 || p > 16 || configuredChunk == 0 || configuredChunk > 16*1024*1024 {
		return errors.New("invalid backup encryption parameters")
	}
	key, err := scrypt.Key([]byte(passphrase), salt, int(n), int(r), int(p), 32)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			_ = os.Remove(outputPath)
		}
	}()

	var expectedIndex uint64
	for {
		index, plainLen, nonce, ciphertext, err := readRecord(reader, gcm.NonceSize(), configuredChunk+uint32(gcm.Overhead()))
		if err != nil {
			return fmt.Errorf("backup is truncated or corrupt: %w", err)
		}
		if index != expectedIndex {
			return errors.New("backup record order is invalid")
		}
		isFinal := plainLen == 0
		plain, err := gcm.Open(nil, nonce, ciphertext, recordAAD(index, plainLen, isFinal))
		if err != nil {
			return errors.New("wrong passphrase or backup integrity check failed")
		}
		if isFinal {
			if len(plain) != 0 {
				return errors.New("invalid backup final record")
			}
			// The authenticated final record is the end of the container. Reject
			// appended bytes so corrupted or polyglot files are not accepted.
			if extra, readErr := reader.ReadByte(); readErr == nil || !errors.Is(readErr, io.EOF) {
				_ = extra
				return errors.New("unexpected trailing data after backup final record")
			}
			break
		}
		if uint32(len(plain)) != plainLen {
			return errors.New("invalid backup record length")
		}
		if _, err := output.Write(plain); err != nil {
			return err
		}
		expectedIndex++
	}
	if err := output.Sync(); err != nil {
		return err
	}
	ok = true
	return nil
}

func writeRecord(writer io.Writer, index uint64, plainLen uint32, nonce, ciphertext []byte) error {
	if err := binary.Write(writer, binary.BigEndian, index); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, plainLen); err != nil {
		return err
	}
	if _, err := writer.Write(nonce); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, uint32(len(ciphertext))); err != nil {
		return err
	}
	_, err := writer.Write(ciphertext)
	return err
}

func readRecord(reader io.Reader, nonceSize int, maxCipher uint32) (uint64, uint32, []byte, []byte, error) {
	var index uint64
	var plainLen uint32
	if err := binary.Read(reader, binary.BigEndian, &index); err != nil {
		return 0, 0, nil, nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &plainLen); err != nil {
		return 0, 0, nil, nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return 0, 0, nil, nil, err
	}
	var cipherLen uint32
	if err := binary.Read(reader, binary.BigEndian, &cipherLen); err != nil {
		return 0, 0, nil, nil, err
	}
	if cipherLen == 0 || cipherLen > maxCipher+64 {
		return 0, 0, nil, nil, errors.New("invalid ciphertext length")
	}
	ciphertext := make([]byte, cipherLen)
	if _, err := io.ReadFull(reader, ciphertext); err != nil {
		return 0, 0, nil, nil, err
	}
	return index, plainLen, nonce, ciphertext, nil
}

func recordAAD(index uint64, plainLen uint32, final bool) []byte {
	buffer := make([]byte, 0, len(magic)+13)
	buffer = append(buffer, magic...)
	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, index)
	buffer = append(buffer, indexBytes...)
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, plainLen)
	buffer = append(buffer, lengthBytes...)
	if final {
		buffer = append(buffer, 1)
	} else {
		buffer = append(buffer, 0)
	}
	return buffer
}
