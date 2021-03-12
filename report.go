package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"sync"
)

var w sync.WaitGroup

// ServerJ server酱通知
func ServerJ(sckey string, text string, desp string) {
	type resMap struct {
		Dataset string `json:"dataset"`
		Errmsg  string `json:"errmsg"`
		Errno   int    `json:"errno"`
	}
	serverJUrl := "https://sc.ftqq.com/" + sckey + ".send?text=" + text + "&desp=" + desp
	resp, err := http.Get(serverJUrl)
	if err != nil {
		logger.Warn("Server酱通知好像失败了:", err.Error())
	}

	respBodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("Server酱通知好像失败了:", err.Error())
	}

	var resBodyMap resMap
	err = json.Unmarshal(respBodyByte, &resBodyMap)
	if err != nil {
		logger.Warn("Server酱通知好像失败了:", err.Error())
	}

	if resBodyMap.Errmsg != "success" {
		logger.Warn("Server酱通知好像失败了:", fmt.Errorf(resBodyMap.Errmsg))

	}
	w.Done()
}

// Pushpluse pushpluse通知
func Pushpluse(token string, title string, content string) {
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
		logger.Warn("Push+通知好像失败了:", err.Error())
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	if err != nil {
		logger.Warn("Push+通知好像失败了:", err.Error())
	}
	clien := &http.Client{}
	res, err := clien.Do(req)
	if err != nil {
		logger.Warn("Push+通知好像失败了:", err.Error())
	}
	resBodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Warn("Push+通知好像失败了:", err.Error())
	}
	var resBodyBlock resBody
	err = json.Unmarshal(resBodyByte, &resBodyBlock)
	if err != nil {
		logger.Warn("Push+通知好像失败了:", err.Error())
	}
	if resBodyBlock.Code != 200 {
		logger.Warn("Push+通知好像失败了:", fmt.Errorf(resBodyBlock.Data))

	}
	w.Done()
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
func SendEmail(to, password, title, body string) {
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
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}
	defer c.Close()

	err = c.Auth(auth)
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}

	err = c.Mail(to)
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}

	err = c.Rcpt(recipients)
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}

	writer, err := c.Data()
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}
	_, err = writer.Write([]byte(smtpMsg))
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}
	err = writer.Close()
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}
	err = c.Quit()
	if err != nil {
		logger.Warn("邮件通知好像失败了: ", err.Error())
	}
	w.Done()
}

// 发送提醒
func (u *User) report(text string) {
	if u.ServerJ != "" {
		w.Add(1)
		go ServerJ(u.ServerJ, "完美校园打卡通知", text)
	}

	if u.Email != "" && u.EmailPassword != "" {
		w.Add(1)
		go SendEmail(u.Email, u.EmailPassword, "完美校园打卡通知", text)
	}

	if u.PushPlus != "" {
		w.Add(1)
		go Pushpluse(u.PushPlus, "完美校园打卡通知", text)
	}
	w.Wait()
}
