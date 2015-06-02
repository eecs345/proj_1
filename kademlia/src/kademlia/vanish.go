package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
    "time"
	mathrand "math/rand"
	"sss"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
    r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
    accessKey = r.Int63()
    return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func VanishData(kadem Kademlia, data []byte, numberKeys byte,
	threshold byte) (vdo VanashingDataObject) {
	return
}

func UnvanishData(kadem Kademlia, vdo VanashingDataObject) (data []byte) {
	Access_Key := vdo.AccessKey
	Cipher_text := vdo.Ciphertext
	Number_Keys := vdo.NumberKeys
	Thres_hold := vdo.Threshold
	ids := CalculateSharedKeyLocations(Access_Key, int64(Number_Keys))
	//need convert ids to shares
	// ids : []ID   shares : map[byte][]byte
	shares := make(map[byte][]byte)
	ids = ids[0:int64(Thres_hold)]
	for _, item := range ids {
		response := kadem.DoIterativeFindValue(item)
		//problem?
		parse_response(response, shares)
	}
	secret := sss.Combine(shares)
	data = decrypt(secret, Cipher_text)

	return
}

func parse_response(response string, shares map[byte][]byte) {
		response = response[163:]
		key := byte(response[0])
		value := []byte(response[1:])
		shares[key] = value
}
