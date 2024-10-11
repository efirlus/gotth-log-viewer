package userinfo

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	directorypath = "/home/efirlus/.config/gobooknotes" // /home/efirlus/ 보안이 가능한 공간 어딘가
	configFile    = "config"
)

/*
INFO: 새로운 알고리즘의 BOOKNOTES에선 사용 안하기로 결정
DB 쓰는 프로그램에서는 쓰게 될거같은데...

	메인 외부 호출 - CheckConfig(strkey) ([]credential{id, pass})
		key := generateHash (strkey)
		if key not = saveCredentials

			encrypt (id), (pass) with key

		else = readCredentials
			decrypt (key) (id, pass)
*/

// config가 있는지 없는지 확인하고, 없으면 저장, 있으면 리딩
func CheckConfig(strkey string) ([]string, error) {
	encryptionKey := generateHash(strkey)
	configFilePath := filepath.Join(directorypath, configFile)

	var user, password string

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// If config file does not exist, prompt for credentials and save them
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter user: ")
		user, _ = reader.ReadString('\n')
		user = strings.TrimSpace(user)

		fmt.Print("Enter password: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)

		err = saveCredentials(user, password, encryptionKey)
		if err != nil {
			errorMessage := fmt.Sprintf("in the func saveCredentials: %v", err)
			return nil, fmt.Errorf("ERROR: %v", errorMessage)
		}
	} else {
		// Read and decrypt credentials from the config file
		user, password, err = readCredentials(encryptionKey)
		if err != nil {
			errorMessage := fmt.Sprintf("in the func readCredentials: %v", err)
			return nil, fmt.Errorf("ERROR: %v", errorMessage)
		}
	}

	return []string{user, password}, nil
}

// 인크립션키를 scrypt 해시로 쓰자
func generateHash(text string) []byte {
	salt := sha256.Sum256([]byte(text)) // Use a SHA-256 hash of the input text as the salt

	// scrypt 함수의 의미    바이트 텍스트, 소금, 시피유 사용량, 블록사이즈, 평행치, 결과물 바이트 사이즈
	// 다시 말해 얘는 32바이트 값이 나옴, 굿
	dk, err := scrypt.Key([]byte(text), salt[:], 32768, 8, 1, 32)
	if err != nil {
		log.Fatal(err)
	}

	return dk
}

// 아이디, 비번을 받아서 암호화한 뒤 config에 저장하는 함수
func saveCredentials(user string, password string, key []byte) error {
	configFilePath := filepath.Join(directorypath, configFile)
	encryptedUser, err := encrypt(user, key)
	if err != nil {
		errorMessage := fmt.Sprintf("while encrypt user name: %v", err)
		return fmt.Errorf("ERROR: %v", errorMessage)
	}
	fmt.Printf("got id (%s) with key (%s) to make env (%s)\n\n", user, key, encryptedUser)

	encryptedPassword, err := encrypt(password, key)
	if err != nil {
		errorMessage := fmt.Sprintf("while encrypt user pass: %v", err)
		return fmt.Errorf("ERROR: %v", errorMessage)
	}
	fmt.Printf("got pass (%s) with key (%s) to make env (%s)\n\n", password, key, encryptedPassword)

	file, err := os.Create(configFilePath)
	if err != nil {
		errorMessage := fmt.Sprintf("while create config file: %v", err)
		return fmt.Errorf("ERROR: %v", errorMessage)
	}
	defer file.Close()

	_, err = file.WriteString(encryptedUser + "\n" + encryptedPassword + "\n")
	errorMessage := fmt.Sprintf("while write user informations to config file: %v", err)
	return fmt.Errorf("ERROR: %v", errorMessage)
}

// AES 인크립션 함수
func encrypt(text string, key []byte) (string, error) {
	// AES 알고리즘에 맞는 사이퍼를 만들고
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		errorMessage := fmt.Sprintf("in the func encrypt -> aes.NewCipher: %v", err)
		return "", fmt.Errorf("ERROR: %v", errorMessage)
	}

	// 암호화 대상 텍스트를 바이트 처리한 뒤
	plainText := []byte(text)
	//바로 사이퍼로 매핑
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]

	// 사이퍼 처리한 텍스트를 랜덤으로 뒤섞어서
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		errorMessage := fmt.Sprintf("in the func encrypt -> io.ReadFull: %v", err)
		return "", fmt.Errorf("ERROR: %v", errorMessage)
	}

	// 암호화처리
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// 암호화 처리된 놈을 다시 base64 인코딩해서 리턴
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// 저장된 config를 읽어내는 함수
func readCredentials(key []byte) (string, string, error) {
	configFilePath := filepath.Join(directorypath, configFile)
	file, err := os.Open(configFilePath)
	if err != nil {
		errorMessage := fmt.Sprintf("while read a config file to get user info: %v", err)
		return "", "", fmt.Errorf("ERROR: %v", errorMessage)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) < 2 {
		return "", "", fmt.Errorf("invalid config file")
	}

	// 키는 함수가 입력받아서 각 줄을 디크립트
	user, err := decrypt(lines[0], key)
	if err != nil {
		errorMessage := fmt.Sprintf("while decrypt user name from a config file: %v", err)
		return "", "", fmt.Errorf("ERROR: %v", errorMessage)
	}

	password, err := decrypt(lines[1], key)
	if err != nil {
		errorMessage := fmt.Sprintf("while decrypt user pass from a config file: %v", err)
		return "", "", fmt.Errorf("ERROR: %v", errorMessage)
	}

	return user, password, nil
}

// AES 디크립션 함수
func decrypt(cryptoText string, key []byte) (string, error) {
	// 일단 base64 디코딩
	cipherText, _ := base64.URLEncoding.DecodeString(cryptoText)

	// 사이퍼를 생성하고
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		errorMessage := fmt.Sprintf("in the func decrypt -> aes.NewCipher: %v", err)
		return "", fmt.Errorf("ERROR: %v", errorMessage)
	}

	// 사이퍼 처리의 역순으로 섞어서
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// 복호화
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	// 그러면 plainText 생성
	return string(cipherText), nil
}
