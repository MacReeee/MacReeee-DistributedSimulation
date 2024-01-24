package crypto

import (
	"encoding/json"
	"os"

	"go.dedis.ch/kyber/v3/sign/bdn"
)

// 定义用于存储的密钥结构
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GenerateAndStoreKeyPairs 生成 n 对密钥对并将它们存储到一个文件中
func GenerateAndStoreKeyPairs(filename string, n int) error {
	var keyPairs []KeyPair

	for i := 0; i < n; i++ {
		privateKey, publicKey := bdn.NewKeyPair(suite, suite.RandomStream())

		// 将密钥序列化为字节
		pubKeyBytes, err := publicKey.MarshalBinary()
		if err != nil {
			return err
		}

		privKeyBytes, err := privateKey.MarshalBinary()
		if err != nil {
			return err
		}

		keyPairs = append(keyPairs, KeyPair{
			PublicKey:  pubKeyBytes,
			PrivateKey: privKeyBytes,
		})
	}

	// 写入文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(keyPairs)
	if err != nil {
		return err
	}

	return nil
}

func LoadKeyPairs(filename string) ([]KeyPair, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keyPairs []KeyPair
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&keyPairs)
	if err != nil {
		return nil, err
	}

	return keyPairs, nil
}
