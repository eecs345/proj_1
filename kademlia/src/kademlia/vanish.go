package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
    "time"
	mathrand "math/rand"
	"sss"
	"fmt"
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
	k := GenerateRandomCryptoKey()
	Ciphertext  := encrypt(k, data)
	m,_  := sss.Split(numberKeys, threshold, k)
	l := GenerateRandomAccessKey()
	ids := CalculateSharedKeyLocations(l, int64(numberKeys))
		// all := make([]byte,0)
		// all = append(all,byte(i))
		// for _,b := range m[byte(i)] {
		// 	all=append(all,b)
		// }
		// kadem.DoStore(&(kadem.SelfContact), ids[i], all)
	i:=0
	for key, val := range m {
		all := make([]byte, 0)
		all = append(all,key)
		all = append(all, val...)
		kadem.DoIterativeStore(ids[i],all)
		i = i+1
	}

	vdo.AccessKey=l
	vdo.Ciphertext=Ciphertext
	vdo.NumberKeys=numberKeys
	vdo.Threshold=threshold

	//kadem.DoStore(&(kadem.SelfContact), key ID, value []byte)
	return vdo
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
	// ids = ids[0:int64(Thres_hold)]
	for _, item := range ids {
		response := kadem.DoIterativeFindValue(item)
		if(response[0] != 'O') {
			continue
		}
		//problem?
		parse_response(response, &shares)
	}
	fmt.Println(int(Thres_hold))
	if(len(shares) < int(Thres_hold)) {
		fmt.Printf("Can not recover data")
		return
	}
	secret := sss.Combine(shares)
	fmt.Println(len(shares))
	data = decrypt(secret, Cipher_text)

	return
}

func parse_response(response string, shares *map[byte][]byte) {
		response = response[46:]
		key := byte(response[0])
		value := []byte(response[1:])
		(*shares)[key] = value
}
