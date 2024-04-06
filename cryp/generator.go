package cryp

import (
	"encoding/json"
	"go.dedis.ch/kyber/v3/pairing"
	"os"

	"go.dedis.ch/kyber/v3/sign/bdn"
)

var (
	suite = pairing.NewSuiteBn256()
	sks   = []string{
		"436babacceaa0e985ce0b06321e50db789a5f9230b5ef3ffddc8ba64d2c896b8",
		"07ba027c422e35ac0b14794f858e924481e873fcb1771f9357e84055d37f580c",
		"80895cd18284111cb38a813c48b8533168b9adefe3ca5f8348261143391a5440",
		"7e079446464688b9f429ee9e6552e127ec413e3edffab7253b1475f7d79a7dc3",
		"6cdda21fb459d22b1616ebff6b90bbc3dc06303d4e9b20cf6f2c843947b340fd",
		"879a78528c8268ee4d4d230620bd74c34fd8568cfaffac15b2cbf30999994b1a",
		"63870376137ff9d1a81fdebc97585afec5c69d89b8171150919bea0e8a86f6c8",
		"2d8ee8b85557b1fd3796411f8c967245c6a5b42725aee9daac864cabe048c929",
		"1fee201786d30274c019dd49c469f532681336179932a278c58c31cca3513373",
		"1bd5284a44bcb26aed34846cf7173eb7093e5d633b5513006cc9bd76928ce500",
	}
	pks = []string{
		"4e5bfe782aa1ddcdf6480170aa710f4ea366ed63f8b711ebf32305fb254470a7633934ce2a3647b58850823cf32da712643afde19b79d24659b5d7d3148a837a4e0fc74c919d39dfad40e41db5df66dcefcf171c26dfb73c181f176d767fa0686702e9b146bcaf2b6905bff1e25dce2078052474f742bf8c42eaf707afdcb3b7",
		"235e29bd16ad97a7ad0ebfd603fd24646720c794ac3acfb5b44fda366a1cb5b878515255f5e6c4550c8652138ab4feedbfa75c6fdd317b9b1bbafd1eae6ddac21422825a80240d787101b93f14fd37bdb63fa8cd0488032f5c87e69f115fda2e5b66422dc4d164a6ff9de252d40c7cca04f51228abce7ff5df40300936452dd8",
		"390d71fa7a3882048041b3d4d96cfb570c0f844b1df3bd3e73115d969e09a42f6d09ffafe980b37b1b1cc976793813f9ea23e82f8000e5badd05d6e66f65631e30793cd8930af7c20909a94e82b190dff7bd5898ee6b864383991e888db9c08b5e8b6f517f5be6e313c1a74b51fabfb1e77cf65ea9e3ed9a834cdeaa101479f6",
		"1118bd064e4f8c40669afa951d8293e2f4a14e12f02a229b6b40f21bb82853d4087dbc712c1a2e22a4047e0f163aa6fb5815a2e0c28ef19755832ac5bc008d4d55959029010376624a00c3ec38fc60501ee16c7bec0459b7b74bb43c3271422a33418b1bc2c7d1770215c2cf65267b7a06947a7276df773cbc81dbc2c4d80938",
		"12d432a8fe2a57ac7aceaf67519ce80ed7f2ab1f9a1811dc2e94394a20d644003affd564337a543797ccf81f49bd051551adfc19bc62f3d2e1a2c3ea32f111a707031a0a5d2f08fa77d6c77eb58e8ed0e299969691153421d3b22aabd5c340720f45dbaee16330668454913bcd62be21895b6e1f86f4dc9698c302901b4a9b87",
		"35d58aa3ff7f8ddb46e5a1c42268f23dc6298e9cbf503dc20caab97e838e186e24318ab659133de1cb82ef8f72919a249529b9f01e840daeadd27b20b9de7b766de34a660dd09a3310ad2f96cb745f723627debced780923da166b8321a6a45e86fc66e2d68408af1abd6de86b9d4c14ba3a82e75de19e14b10ddf2e6ab2b9e2",
		"06e91e4c752aff38f82750bd797e071fa08861fb35cb079b3385a58535201c0e385e683f03d7d3506170b4b594d1cfc43988debc9103002b6759cc99f0241e6850173d7a23b3b28ae76f0d15ddc9c756cbc51e7b639cb23f921c0fd9ff0e5a8654127e259e5340c877260be46165f09e7de765c7cb0fb97a726827ca940e475a",
		"8317dfe0d2a6702037a4bffe4e062e3dcd4c3b78d633d0e48bf8295c0df263293bd67139e35b8dfdeb63790c14005c2da45993811b8fdd1527ca37be3fd2298f524dbea9e94883853eb7b15f677f52156089590b82171e2d23e026ff0bdc73d64a19878c8753ce3eafef54952115b60043a7deddeb19cccb9a0e3bc469d4cfb5",
		"6ad0ecd79fa0bfab9b757ded51ff9a3b0f75751f6ffd539ceff586c8ef7a336042ed2fd900674488d2b4e88d2e4f341d0deafc290e29b1aba90e2ab7e6abc63d58815136d77f67d134dccf6b21acd45753e0d0d0b75a1cbb86dcb430dcf5b26d825ef752a570f3af07d507b60324dd4423a1e9d818826838a66828ae598a9138",
		"47be4a0c90e9277f8bdd57faa7a747c918ca6543295ef3d4d90c378dc9576cb34759d345dd5b6a33b077a11c8d53acc52a412d277c508f6d28f5d9fe80239e485313f681de7e98dfb0f0869ebda39956fe49ceb65d913ff1ae348ef2f46c41e58564e3e27c36d89b2d34196ecf5988f5fe52c9ff5116898a08a3f94695e4a9bb",
	}
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
