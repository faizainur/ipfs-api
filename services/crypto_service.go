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

// CryptoService represent the object of cryptographic service
// It has methods use for performing cryptographic mechanism
// Such as encryption, decryption, etc.
// Supported algorithm : AES
type CryptoService struct {
	secret     []byte            // Master key
	collection *mongo.Collection // MongoDB crypto collection
}

// UserKey represent the cryptographic key for each user
// Each user key is stored in MongoDB with the user's email address
// With this each key is tied to only one key
type UserKey struct {
	Email string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"`
	Key   string `json:"key,omitempty"  bson:"key"  form:"key"  binding:"key"`
}

// NewCryptoService creates and return a new instance of CrypoService object
// Takes "secret" from parameter as a master key for each cryptographic operation
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

// AESEncrypt perform encryption operation using AES algorithm
// Takes key as cryptographic key used for encryption operation
// return encrypted data in bytes
func (c *CryptoService) AESEncrypt(key []byte, data []byte) []byte {
	// Craete a new AES cipher
	block, errChiper := aes.NewCipher(key)
	if errChiper != nil {
		panic(errChiper)
	}

	// Set cipher mode to GCM
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

// AESDecrypt perform decryption operation using AES algorithm
// Takes key as cryptographic key used for decryption operation
// returns decrypted data in bytes
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

// EncryptUserFile perform encryption to a given file
// Takes two parameters, user email address and file data in bytes
// User email address used for fetching user's cryptographic key from database
// This key will be used for performing encryption operation for a given file data
// returns encrypted data represented in bytes
func (c *CryptoService) EncryptUserFile(email string, file []byte) ([]byte, error) {
	// Fetch cryptographic key based on the email address
	key := c.FetchDecryptedKey(email)
	if key != nil {
		// User key exist
		// Encrypt file using the existing key
		encryptedFile := c.AESEncrypt(key, file)
		return encryptedFile, nil
	}
	// Key is not exist for the given email
	// Generate a new user key and store the encrypted key to database
	// Encrypted user key is encrypted using master key
	key, err := c.GenerateUserKeyWithStoring(email)
	if err != nil {
		return nil, err
	}
	encryptedFile := c.AESEncrypt(key, file)
	return encryptedFile, nil
}

func (c *CryptoService) DecryptUserFile(email string, file []byte) []byte {
	// Fetch cryptographic key based on the email address
	key := c.FetchDecryptedKey(email)
	if key != nil {
		// User key exist
		// Decrypt file using the existing key
		decryptedFile := c.AESDecrypt(key, file)
		return decryptedFile
	}
	// Key is not exist
	// Cannot decrypt file
	return nil
}

// FetchKey used for fetching user key for a given email address
// return cryptographic key in bytes
func (c *CryptoService) FetchKey(email string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var data UserKey
	filter := bson.D{{"email", email}}

	// Find key from database, bind the data to UserKey object
	errMongo := c.collection.FindOne(ctx, filter).Decode(&data)
	if errMongo != nil {
		// key is not exist for this email
		return nil
	}
	// Decode key from hex to retrieve decoded key in bytes
	key, err := hex.DecodeString(data.Key)
	if err != nil {
		panic(err)
	}
	return key
}

// FetchDecryptedKey fetch user key from database
// And decrypt fetched key to get the original user key
// return cryptographic key bytes
func (c *CryptoService) FetchDecryptedKey(email string) []byte {
	// Fetch key for a given email
	key := c.FetchKey(email)
	if key == nil {
		// Key for the given email is not exist
		return nil
	}
	// Decrypt key using the master key
	// to retrieve the original user key
	decryptedKey := c.AESDecrypt(c.secret, key)
	return decryptedKey
}

// StoreKey function store the user key to the database
// User key is stored together with the user email address
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

// GenerateUserKeyWithStoring generates a new user cryptographic key
// encrypted the generated key using the master key
// and encode the encrypted key to hex
// also store the key in to the database together with the user email address
// return unencrypted key in bytes
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

// IsEmailExist check if the email address is registered in the database
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
