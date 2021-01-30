package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/FNDHSTD/logor"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

// Config 配置文件
type Config struct {
	Users []User `json:"users"`
}

// 加载配置文件
func loadConfig() (Config, error) {
	var config Config
	fData, err := ioutil.ReadFile("config.json")
	if err != nil {
		return config, fmt.Errorf("配置文件无法打开, %v", err)
	}
	err = json.Unmarshal(fData, &config)
	if err != nil {
		return config, fmt.Errorf("配置文件解析失败, %v", err)
	}
	return config, nil
}

// User 用户结构体，存储所有的初始信息
type User struct {
	// Session 就是打卡要用的 token,但是需要激活才行
	Session       string
	Username      string `json:"username"`
	Password      string `json:"passworld"`
	Email         string `json:"email"`
	EmailPassword string `json:"emailPassword"`
	PrivateKey    *rsa.PrivateKey
	Key           string
	DeviceID      string `json:"deviceId"`
	CheckData     string
}

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

// rsa加密
func (u *User) rsaEncrypt(input string) ([]byte, error) {
	output, err := rsa.EncryptPKCS1v15(rand.Reader, &u.PrivateKey.PublicKey, []byte(input))
	if err != nil {
		return output, err
	}
	return output, nil
}

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

// des3解码
func (u *User) des3Decrypt(input []byte) ([]byte, error) {
	cipherBlk, err := des.NewTripleDESCipher([]byte(u.Key))
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(cipherBlk, []byte("66666666"))
	output := make([]byte, len(input))
	blockMode.CryptBlocks(output, input)
	output = PKCS5UnPadding(output)
	return output, nil
}

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
	loginArgsMap["userName"] = u.Username
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

// 获取上一次的打卡信息
func (u *User) getLastCheckInData() error {
	// 准备上传的数据
	jsonDataOfUploadJSONMap := make(map[string]string)
	jsonDataOfUploadJSONMap["templateid"] = "pneumonia"
	jsonDataOfUploadJSONMap["token"] = u.Session

	uploadJSONMap := make(map[string]interface{})
	uploadJSONMap["businessType"] = "epmpics"
	uploadJSONMap["method"] = "userComeApp"
	uploadJSONMap["jsonData"] = jsonDataOfUploadJSONMap
	// uploadJSONByte就是要上传的json数据
	uploadJSONByte, err := json.Marshal(uploadJSONMap)

	// 初始化一个请求对象
	req, err := http.NewRequest("POST", "https://reportedh5.17wanxiao.com/sass/api/epmpics", bytes.NewBuffer(uploadJSONByte))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 5.1.1; vmos Build/LMY48G; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Mobile Safari/537.36 Wanxiao/5.3.4")
	req.Header.Set("content-type", "application/json;charset\u003dUTF-8")
	req.Header.Set("x-requested-with", "com.newcapec.mobile.ncp")

	// Do it!
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// 获取服务器返回的数据
	resBodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resBodyMap := make(map[string]string)
	err = json.Unmarshal(resBodyByte, &resBodyMap)
	if err != nil {
		return err
	}
	data := resBodyMap["data"]

	dataMap := make(map[string]interface{})
	json.Unmarshal([]byte(data), &dataMap)

	u.CheckData = fmt.Sprintf(
		`{
			"businessType": "epmpics",
			"method": "submitUpInfo",
			"jsonData": {
				"deptStr": {
					"deptid": %v,
					"text": "%v"
				},
				"areaStr": "%v",
				"reportdate": %v,
				"customerid": "%v",
				"deptid": %v,
				"source": "app",
				"templateid": "%v",
				"stuNo": "%v",
				"username": "%v",
				"phonenum": "",
				"userid": "%v",
				"updatainfo": [
					{
						"propertyname": "temperature",
						"value": "36.4"
					},
					{
						"propertyname": "symptom",
						"value": "无症状"
					},
					{
						"propertyname": "isConfirmed",
						"value": "否"
					},
					{
						"propertyname": "isdefinde",
						"value": "否.未隔离"
					},
					{
						"propertyname": "isTouch",
						"value": "否"
					},
					{
						"propertyname": "isTransitArea",
						"value": "否"
					},
					{
						"propertyname": "是否途径或逗留过疫情中，高风险地区？",
						"value": ""
					},
					{
						"propertyname": "isFFHasSymptom",
						"value": "没有"
					},
					{
						"propertyname": "isContactFriendIn14",
						"value": "没有"
					},
					{
						"propertyname": "xinqing",
						"value": "健康"
					},
					{
						"propertyname": "bodyzk",
						"value": "是"
					},
					{
						"propertyname": "cxjh",
						"value": "否"
					},
					{
						"propertyname": "isleaveaddress",
						"value": "否"
					},
					{
						"propertyname": "isAlreadyInSchool",
						"value": "没有"
					},
					{
						"propertyname": "ownPhone",
						"value": "%v"
					},
					{
						"propertyname": "emergencyContact",
						"value": "%v"
					},
					{
						"propertyname": "mergencyPeoplePhone",
						"value": "%v"
					},
					{
						"propertyname": "assistRemark",
						"value": ""
					}
				],
				"gpsType": 1,
				"token": "%v"
			}
		}`,
		dataMap["deptStr"].(map[string]interface{})["deptid"].(float64),
		dataMap["deptStr"].(map[string]interface{})["text"].(string),
		// dataMap["areaStr"].(string),
		strings.ReplaceAll(dataMap["areaStr"].(string), `"`, `\"`),
		time.Now().UnixNano()/1e6,
		dataMap["customerid"],
		dataMap["deptStr"].(map[string]interface{})["deptid"].(float64),
		dataMap["templateid"],
		dataMap["stuNo"],
		dataMap["username"],
		dataMap["userid"],
		dataMap["cusTemplateRelations"].([]interface{})[14].(map[string]interface{})["value"].(string),
		dataMap["cusTemplateRelations"].([]interface{})[15].(map[string]interface{})["value"].(string),
		dataMap["cusTemplateRelations"].([]interface{})[16].(map[string]interface{})["value"].(string),
		u.Session,
	)

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
		return fmt.Errorf("Session激活失败, %v", msg)
	}

	return nil
}

// 打卡
func (u *User) checkIn() error {
	// checkData := strings.ReplaceAll(u.CheckData, "\n", "")
	checkData := strings.ReplaceAll(strings.ReplaceAll(u.CheckData, "\n", ""), "\t", "")
	// 初始化一个请求对象
	req, err := http.NewRequest("POST", "https://reportedh5.17wanxiao.com/sass/api/epmpics", bytes.NewBuffer([]byte(checkData)))
	if err != nil {
		return nil
	}
	// 设置请求头
	req.Header.Set("Origin", "https://reportedh5.17wanxiao.com")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "com.eg.android.AlipayGphone")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 5.1.1; vmos Build/LMY48G; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Mobile Safari/537.36 Wanxiao/5.3.4")
	req.Header.Set("x-mass-tappid", "2019030163398604")
	req.Header.Set("Content-Type", "application/json;charset\u003dUTF-8")
	req.Header.Set("Host", "reportedh5.17wanxiao.com")
	req.Header.Set("Connection", "Keep-Alive")
	// Do it
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	resBodyByte, err := ioutil.ReadAll(res.Body)
	resBodyMap := make(map[string]interface{})
	err = json.Unmarshal(resBodyByte, &resBodyMap)
	if err != nil {
		return nil
	}
	msg, ok := resBodyMap["msg"].(string)
	if !ok {
		return fmt.Errorf("服务器返回数据异常")
	}
	if msg != "成功" {
		data, ok := resBodyMap["data"].(string)
		if !ok {
			data = "服务器返回数据异常"
		}
		return fmt.Errorf("登录失败, %v", data)
	}
	return nil
}

func dial(addr string) (*smtp.Client, error) {
	connect, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Println("Dial:", err)
		return nil, err
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return smtp.NewClient(connect, host)
}

func (u *User) sendEmail(to, password, title, body string) error {
	// 设置邮件内容头部信息
	header := make(map[string]string)
	header["From"] = "WanxiaoHealthyCheck"
	header["TO"] = to
	header["Subject"] = title
	header["Content-Type"] = "text/html;chartset=UTF-8"

	// 将头部信息拼接
	var smtpMsg string
	for k, v := range header {
		smtpMsg += k + ":" + v + "\r\n"
	}

	// 将正文拼接
	smtpMsg += "\r\n" + body

	// 初始化一个作者变量
	// auth := smtp.PlainAuth("", "1713252605@qq.com", "yfktavpmfbxvbfjc", "smtp.qq.com")
	auth := smtp.PlainAuth("", to, password, "smtp.qq.com")

	recipients := to

	c, err := dial("smtp.qq.com:465")
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.Auth(auth)
	if err != nil {
		return err
	}

	err = c.Mail(to)
	if err != nil {
		return err
	}

	err = c.Rcpt(recipients)
	if err != nil {
		return err
	}

	writer, err := c.Data()
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(smtpMsg))
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	err = c.Quit()
	if err != nil {
		return err
	}
	return nil
}

func wanxiaoHealthyCheck() {
	logger := logor.NewConsoleLogger("debug")
	config, err := loadConfig()
	if err != nil {
		logger.Error("读取配置文件失败：%v", err)
		return
	}
	for _, user := range config.Users {
		err = user.createRSAKey()
		if err != nil {
			logger.Error("用户%v生成密钥失败：%v", user.Username, err)
			return
		}
		err = user.exchangeKey()
		if err != nil {
			logger.Error("与服务器交换得到的密钥失败, %v", err)
			return
		}
		logger.Info("1. 与服务器交换得到的密钥key: %v和session:%v", user.Key, user.Session)
		err := user.login()
		if err != nil {
			logger.Error("登录失败: %v", err)
			return
		}
		logger.Info("2. 登录成功！")
		err = user.activateSession()
		if err != nil {
			logger.Error("Session激活失败, %v", err)
			return
		}
		logger.Info("3. Session激活为Token成功！")
		// 获取上一次的打卡信息
		err = user.getLastCheckInData()
		if err != nil {
			logger.Error("获取上次打卡信息失败, %v", err)
			return
		}
		// 打卡
		err = user.checkIn()
		if err != nil {
			logger.Error("用户%v打卡失败, err:%v", user.Username, err)
			err := user.sendEmail(user.Email, user.EmailPassword, "完美校园打卡通知", "用户"+user.Username+"今日打卡失败")
			if err != nil {
				logger.Error("用户%v邮件发送失败, err:%v", user.Username, err)
			}
			return
		}
		logger.Info("4. 用户%v打卡成功", user.Username)
		err = user.sendEmail(user.Email, user.EmailPassword, "完美校园打卡通知", "用户"+user.Username+"今日打卡成功")
		if err != nil {
			logger.Error("用户%v邮件发送失败, err:%v", user.Username, err)
		}
		logger.Info("5. 用户%v邮件发送成功\n", user.Username)
	}
}

func main() {
	cloudfunction.Start(wanxiaoHealthyCheck)
}
