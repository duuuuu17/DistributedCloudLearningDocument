package my_utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"gintest/types"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/yaml.v3"
)

var contentFile map[string]interface{} = make(map[string]interface{})

type MyClaim struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var filename string = "properties.yml"

func InitKey() {
	log.Printf("JWT Initialization !\n")
	CreatePrivateKeyFromPropertiesYAML()
	log.Println("JWT initial finished !")
}

func ReadPrivateKey() {
	// os.ReadFile
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Failure read from file: ", err)
		return
	}
	// yaml.Unmarshal
	if err := yaml.Unmarshal(yamlFile, &contentFile); err != nil {
		CreatePrivateKeyFromPropertiesYAML()
		return
	}
	contentFile["PrivateKey"] = StringToBytes(contentFile["PrivateKey"].(string))
	contentFile["PublicKey"] = StringToBytes(contentFile["PublicKey"].(string))
}

// x509 package that convert key to DER format []byte's array
// create PEM file and save the file to ContentFile's that a object type of map[string]interface{}
func CreatePrivateKeyFromPropertiesYAML() {
	// rsa.generatKey && encode x509's certification ,got pem file
	priaveKey, _ := rsa.GenerateKey(rand.Reader, 2048) // generate rsa's key object
	// PKCS格式编码私钥生成的PEM更原始，仅支持RSA算法
	derStream := x509.MarshalPKCS1PrivateKey(priaveKey) // encode private key that a x509 certification
	block := &pem.Block{                                // PEM Block
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey := pem.EncodeToMemory(block)        // got privateKey
	contentFile["PrivateKey"] = string(prvKey) // write privateKey

	publicKey := &priaveKey.PublicKey // get privatekey (pem)
	// 因为公钥往往可能需要在网络上传播，并且支持多种算法：RSA、ECDSA等
	derPkix, _ := x509.MarshalPKIXPublicKey(publicKey) // encode public key that a x509 certification
	block = &pem.Block{                                // PEM block
		Type:  "RSA PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey := pem.EncodeToMemory(block)       // got publicKey (pem)
	contentFile["PublicKey"] = string(pubKey) // write publicKey

	// Marshal YAML content && Rewrite
	newYaml, err := yaml.Marshal(contentFile)
	if err != nil {
		log.Fatal("Failure encode content: ", err)
	}
	if err := os.WriteFile(filename, newYaml, 0644); err != nil {
		log.Fatalf("write file failure! error: %v\n", err)
	}
	// 写入文件完毕后，将其对应内容修改为[]byte类型
	contentFile["PrivateKey"] = prvKey
	contentFile["PublicKey"] = pubKey
}

// get privateKey from ContentFile object and the objet save privateKey&publicKey
func GetPrivateKey(user *types.LoginUser) bool {
	if _, exists := contentFile["PrivateKey"]; exists {
		CreatePrivateKeyFromPropertiesYAML()
		return false
	}
	return true
}

func GenerateJWT(user *types.LoginUser) string {
	// generate ID Number
	randomNum := make([]byte, 32)
	rand.Read(randomNum)
	h := sha256.Sum256(randomNum)
	idNum := binary.BigEndian.Uint32(h[:])
	// create json jwtClaim
	claims := MyClaim{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "WebServer",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  jwt.ClaimStrings{user.Username},
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			ID:        string(rune(idNum)),
		},
	}
	// got token && string，指定签名算法要用RSA类型
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// 使用PEM格式的私钥进行签名
	privateKeyData, _ := jwt.ParseRSAPrivateKeyFromPEM(contentFile["PrivateKey"].([]byte))
	signalString, _ := token.SignedString(privateKeyData)
	if signalString == "" {
		log.Fatalln("the token was generated null!")
		return ""
	}
	return signalString
}
func ParseTokenCheckExpired(token string) bool {
	if token == "" {
		return false
	}
	// 此处要注意，在最开始生成PEM时，使用的编码格式与此处公钥解析时使用的算法相同（jwt公钥使用的PKI）才可以解析出内容
	// publicKey, _ := pem.Decode(contentFile["PublicKey"].([]byte))
	// pubKey, err := x509.ParsePKIXPublicKey(publicKey.Bytes)
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(contentFile["PublicKey"].([]byte))
	if err != nil {
		log.Fatalln("got publicKey error: ", err)
	}
	// KeyFunc 用于返回验证JWT签名的密钥
	myClaim := &MyClaim{}
	claim, err := jwt.ParseWithClaims(token, myClaim, func(t *jwt.Token) (interface{}, error) {
		return pubKey, nil
	})
	ErrorInLog(err)
	// 验证token是否有效
	if _, ok := claim.Claims.(*MyClaim); ok && claim.Valid {
		return false
	}
	log.Println("it's not valid")
	return true
}
