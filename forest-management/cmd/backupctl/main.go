package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"forest-management/pkg/backupcrypto"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "decrypt" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/backupctl decrypt -in backup.bsmbackup -out backup.zip [-passphrase-file /secure/path]")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("decrypt", flag.ExitOnError)
	input := flags.String("in", "", "encrypted backup path")
	output := flags.String("out", "", "decrypted output path")
	passphraseFile := flags.String("passphrase-file", "", "optional path to a 0600 passphrase file")
	_ = flags.Parse(os.Args[2:])
	if *input == "" || *output == "" {
		log.Fatal("-in and -out are required")
	}
	passphrase, err := readPassphrase(*passphraseFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := backupcrypto.DecryptFile(*input, *output, passphrase); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Backup decrypted successfully")
}

func readPassphrase(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("read passphrase file: %w", err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("passphrase file must be a regular file")
		}
		// On Unix, reject files readable by group or others. Windows does not
		// expose equivalent permission bits through FileMode, so deployment ACLs
		// must be verified separately there.
		if info.Mode().Perm() != 0 && info.Mode().Perm()&0o077 != 0 {
			return "", fmt.Errorf("passphrase file permissions are too broad; use 0600")
		}
		value, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read passphrase file: %w", err)
		}
		passphrase := strings.TrimSpace(string(value))
		if len(passphrase) < 16 {
			return "", fmt.Errorf("backup passphrase must contain at least 16 characters")
		}
		return passphrase, nil
	}

	fmt.Print("Backup passphrase (input may be echoed; prefer -passphrase-file): ")
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	passphrase := strings.TrimSpace(value)
	if len(passphrase) < 16 {
		return "", fmt.Errorf("backup passphrase must contain at least 16 characters")
	}
	return passphrase, nil
}
