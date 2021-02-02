package report

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
)

// ServerJ server酱通知
func ServerJ(sckey string, text string, desp string) error {
	type resMap struct {
		Dataset string `json:"dataset"`
		Errmsg  string `json:"errmsg"`
		Errno   int    `json:"errno"`
	}
	serverJUrl := "https://sc.ftqq.com/" + sckey + ".send?text=" + text + "&desp=" + desp
	resp, err := http.Get(serverJUrl)
	if err != nil {
		return err
	}

	respBodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var resBodyMap resMap
	err = json.Unmarshal(respBodyByte, &resBodyMap)
	if err != nil {
		return err
	}

	if resBodyMap.Errmsg != "success" {
		return fmt.Errorf(resBodyMap.Errmsg)
	}

	return nil
}

// Pushpluse pushpluse通知
func Pushpluse(token string, title string, content string) error {
	type resBody struct {
		Code int    `json:"code"`
		Data string `json:"data"`
		Msg  string `json:"msg"`
	}
	url := "http://pushplus.hxtrip.com/send/"
	jsonMap := make(map[string]string)
	jsonMap["token"] = token
	jsonMap["title"] = title
	jsonMap["content"] = content
	jsonByte, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	clien := &http.Client{}
	res, err := clien.Do(req)
	if err != nil {
		return err
	}
	resBodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var resBodyBlock resBody
	err = json.Unmarshal(resBodyByte, &resBodyBlock)
	if resBodyBlock.Code != 200 {
		return fmt.Errorf(resBodyBlock.Data)
	}
	return nil
}

// 发起TLS
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

// SendEmail 发送邮件
func SendEmail(to, password, title, body string) error {
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
