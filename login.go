package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 生成 RSA 密钥
func (u *User) createRSAKey() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return fmt.Errorf("生成RSA密钥失败, %v", err)
	}
	u.PrivateKey = privateKey
	return nil
}

// 与服务器交换密钥
func (u *User) exchangeKey() error {
	x509PublicKey, err := x509.MarshalPKIXPublicKey(&u.PrivateKey.PublicKey)
	if err != nil {
		return err
	}
	stringPublicKey := base64.RawStdEncoding.EncodeToString(x509PublicKey)
	req, err := http.NewRequest("POST", "https://server.17wanxiao.com/campus/cam_iface46/exchangeSecretkey.action", bytes.NewBuffer([]byte(fmt.Sprintf(`{"key":"%v"}`, stringPublicKey))))

	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://server.17wanxiao.com")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 读取服务器传回的数据
	resbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// base64解码
	jsonBase64, err := base64.StdEncoding.DecodeString(string(resbody))
	if err != nil {
		return err
	}

	// rsa解码
	jsonRsa, err := u.rsaDecrypt(string(jsonBase64))
	if err != nil {
		return err
	}
	// fmt.Println("jsonRsa", jsonRsa)
	jsonMap := make(map[string]string)
	err = json.Unmarshal([]byte(jsonRsa), &jsonMap)
	if err != nil {
		return err
	}
	u.Key = jsonMap["key"][:24]
	u.Session = jsonMap["session"]
	return nil
}

// rsa解密
func (u *User) rsaDecrypt(input string) (string, error) {
	output, err := rsa.DecryptPKCS1v15(rand.Reader, u.PrivateKey, []byte(input))
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

// // rsa加密
// func (u *User) rsaEncrypt(input string) ([]byte, error) {
// 	output, err := rsa.EncryptPKCS1v15(rand.Reader, &u.PrivateKey.PublicKey, []byte(input))
// 	if err != nil {
// 		return output, err
// 	}
// 	return output, nil
// }

// PKCS5Padding 填充
func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// PKCS5UnPadding 取消填充
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// des3编码
func (u *User) des3Encrypt(input []byte) ([]byte, error) {
	cipherBlk, err := des.NewTripleDESCipher([]byte(u.Key))
	if err != nil {
		return nil, err
	}
	input = PKCS5Padding(input, cipherBlk.BlockSize())
	blockMode := cipher.NewCBCEncrypter(cipherBlk, []byte("66666666"))
	output := make([]byte, len(input))
	blockMode.CryptBlocks(output, input)
	return output, nil
}

// // des3解码
// func (u *User) des3Decrypt(input []byte) ([]byte, error) {
// 	cipherBlk, err := des.NewTripleDESCipher([]byte(u.Key))
// 	if err != nil {
// 		return nil, err
// 	}
// 	blockMode := cipher.NewCBCDecrypter(cipherBlk, []byte("66666666"))
// 	output := make([]byte, len(input))
// 	blockMode.CryptBlocks(output, input)
// 	output = PKCS5UnPadding(output)
// 	return output, nil
// }

// 计算sha256
func getSha256(input []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(input)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), err
}

// 模拟登录
func (u *User) login() error {
	// 对密码进行des3加密
	passwordDES3, err := u.des3Encrypt([]byte(u.Password))
	if err != nil {
		return err
	}
	// 对密文进行 base64 编码
	password := base64.StdEncoding.EncodeToString(passwordDES3)
	// 准备要上传的数据
	loginArgsMap := make(map[string]interface{})
	loginArgsMap["appCode"] = "M002"
	loginArgsMap["deviceId"] = u.DeviceID
	loginArgsMap["netWork"] = "wifi"
	// tmd，password字段必须是列表形式，否则会返回一大堆jsva错误，在fastjson包中提示语法错误
	loginArgsMap["password"] = []string{password}
	loginArgsMap["qudao"] = "guanwang"
	loginArgsMap["requestMethod"] = "cam_iface46/loginnew.action"
	loginArgsMap["shebeixinghao"] = ""
	loginArgsMap["systemType"] = "android"
	loginArgsMap["telephoneInfo"] = "5.1.1"
	loginArgsMap["telephoneModel"] = ""
	loginArgsMap["type"] = "1"
	loginArgsMap["userName"] = u.PhoneNumber
	loginArgsMap["wanxiaoVersion"] = 10462101
	loginArgsMap["yunyingshang"] = "07"
	// 将mapjson序列化
	loginArgsJSON, err := json.Marshal(loginArgsMap)
	if err != nil {
		return err
	}
	// 对loginArgsJSON进行des3加密
	loginArgsJSON, err = u.des3Encrypt(loginArgsJSON)
	if err != nil {
		return err
	}
	// 对loginArgsJSON进行base64编码
	loginArgsJSONStr := base64.StdEncoding.EncodeToString(loginArgsJSON)
	// 准备直接上传的结构体
	jsonMap := make(map[string]string)
	jsonMap["session"] = u.Session
	jsonMap["data"] = string(loginArgsJSONStr)

	// jsonByte 要直接上传的json数据
	jsonByte, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://server.17wanxiao.com/campus/cam_iface46/loginnew.action", bytes.NewBuffer(jsonByte))
	if err != nil {
		return err
	}
	// 计算sha256
	jsonByte256, err := getSha256(jsonByte)
	if err != nil {
		return err
	}

	req.Header.Set("campusSign", jsonByte256)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	resMap := make(map[string]interface{})
	err = json.Unmarshal(resbody, &resMap)
	if err != nil {
		return err
	}
	resul, ok := resMap["result_"].(bool)
	if !ok {
		return fmt.Errorf("服务器返回数据异常")
	}
	if !resul {
		msg, ok := resMap["message_"].(string)
		if !ok {
			return fmt.Errorf("登录失败,服务器返回数据异常")
		}
		return fmt.Errorf("登录失败%v", msg)
	}
	return nil
}

// 激活Session
func (u *User) activateSession() error {
	postStr := "appClassify=DK&token=" + u.Session
	req, err := http.NewRequest("POST", "https://reportedh5.17wanxiao.com/api/clock/school/getUserInfo", bytes.NewBuffer([]byte(postStr)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	// 读取数据
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// 解析数据
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}
	result, ok := bodyMap["result"].(bool)
	if !ok {
		return fmt.Errorf("服务器返回数据异常, result类型断言失败")
	}
	if !result {
		msg, ok := bodyMap["msg"].(string)
		if !ok {
			return fmt.Errorf("服务器返回数据异常, result类型断言失败")
		}
		return fmt.Errorf("session激活失败, %v", msg)
	}

	return nil
}
