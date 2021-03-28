package cutils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func GenerateKey() []byte {
	key := make([]byte, 24)

	_, err := rand.Read(key)
	if err != nil {
		log.Fatal(err.Error())
	}
	return key
}

func GenerateKeyFile() []byte {
	key := GenerateKey()
	encodedKey := hex.EncodeToString(key)

	dirPath := GetKeyDirPath()
	os.MkdirAll(dirPath, 0744)

	out, err := os.Create(fmt.Sprintf("%s/%s", dirPath, "master.key"))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()

	_, errWrite := out.WriteString(encodedKey)
	if errWrite != nil {
		log.Fatal(errWrite.Error())
	}
	return key
}

func GetKeyDirPath() string {
	homeDIr, errHomeDir := os.UserHomeDir()
	if errHomeDir != nil {
		log.Fatal(errHomeDir.Error())
	}
	dirPath := fmt.Sprintf("%s/%s", homeDIr, ".catena")
	return dirPath
}

func GetKeyPath() string {
	return fmt.Sprintf("/%s/%s", GetKeyDirPath(), "master.key")
}
