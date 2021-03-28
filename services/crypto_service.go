package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/faizainur/ipfs-api/cutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CryptoService struct {
	secret     []byte
	collection *mongo.Collection
}

type UserKey struct {
	Email string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"`
	Key   string `json:"key,omitempty"  bson:"key"  form:"key"  binding:"key"`
}

func NewCryptoService(secret []byte) *CryptoService {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbConfig := cutils.DbUtils{ConnectionString: os.Getenv("MONGODB_URI")}
	client, err := dbConfig.Connect(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Connected to Mongo URI")

	dbCrypto := client.Database("crypto")
	collection := dbCrypto.Collection("secret")

	return &CryptoService{
		secret:     secret,
		collection: collection,
	}
}

func (c *CryptoService) AESEncrypt(key []byte, data []byte) []byte {
	block, errChiper := aes.NewCipher(key)
	if errChiper != nil {
		panic(errChiper)
	}

	gcm, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		panic(errGcm)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	chipertext := gcm.Seal(nonce, nonce, data, nil)

	return chipertext
}

func (c *CryptoService) AESDecrypt(key []byte, data []byte) []byte {
	block, errChiper := aes.NewCipher(key)
	if errChiper != nil {
		panic(errChiper)
	}

	gcm, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		panic(errGcm)
	}

	nonce := data[:gcm.NonceSize()]
	chipertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, chipertext, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}

func (c *CryptoService) EncryptUserFile(email string, file []byte) ([]byte, error) {
	// key := make([]byte, 24)

	key := c.FetchDecryptedKey(email)
	if key != nil {
		// User key exist
		// Encrypt file with the existing key
		encryptedFile := c.AESEncrypt(key, file)
		return encryptedFile, nil
	}
	// Key is not exist for the email
	// Generate new user key and store the encrypted key to database
	// Encrypted user key is encrypted using master key
	key, err := c.GenerateUserKeyWithStoring(email)
	if err != nil {
		return nil, err
	}
	encryptedFile := c.AESEncrypt(key, file)
	return encryptedFile, nil
}

func (c *CryptoService) DecryptUserFile(email string, file []byte) []byte {
	key := c.FetchDecryptedKey(email)
	if key != nil {
		// User key exist
		// Decrypt file with the existing key
		decryptedFile := c.AESDecrypt(key, file)
		return decryptedFile
	}
	// Key is not exist
	// Cannot decrypt file
	return nil
}

func (c *CryptoService) FetchKey(email string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var data UserKey

	filter := bson.D{{"email", email}}

	errMongo := c.collection.FindOne(ctx, filter).Decode(&data)
	if errMongo != nil {
		// key not exist for this email
		return nil
	}

	key, err := hex.DecodeString(data.Key)
	if err != nil {
		panic(err)
	}
	return key
}

func (c *CryptoService) FetchDecryptedKey(email string) []byte {
	key := c.FetchKey(email)
	if key == nil {
		return nil
	}
	decryptedKey := c.AESDecrypt(c.secret, key)
	return decryptedKey
}

func (c *CryptoService) StoreKey(email string, key string) error {
	var userKey UserKey

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	isExist, errEmailExist := c.IsEmailExist(email)
	if errEmailExist != nil {
		panic(errEmailExist)
	}

	if isExist {
		return nil
	}

	userKey.Email = email
	userKey.Key = key

	_, err := c.collection.InsertOne(ctx, userKey)
	if err != nil {
		return err
	}
	return nil
}

func (c *CryptoService) GenerateUserKeyWithStoring(email string) ([]byte, error) {
	key := cutils.GenerateKey()

	encryptedKey := c.AESEncrypt(c.secret, key)

	encodedKey := hex.EncodeToString(encryptedKey)
	err := c.StoreKey(email, string(encodedKey))
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (c *CryptoService) IsEmailExist(email string) (bool, error) {
	var isExist bool = false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetMaxTime(2 * time.Second)
	count, err := c.collection.CountDocuments(ctx, bson.D{{"email", email}}, opts)

	if count > 0 && err == nil {
		isExist = true
	}

	return isExist, err
}
