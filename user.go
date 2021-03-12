package main

import "crypto/rsa"

// User 用户结构体，存储所有的初始信息
type User struct {
	PhoneNumber   string `toml:"phoneNumber"`
	Password      string `toml:"password"`
	DeviceID      string `toml:"deviceID"`
	Email         string `toml:"email"`
	EmailPassword string `toml:"emailPassword"`
	ServerJ       string `toml:"serverJ"`
	PushPlus      string `toml:"pushPlus"`
	HealthyCheck  bool   `toml:"healthyCheck"`
	InSchoolCheck bool   `toml:"inSchoolCheck"`
	StdNo         string
	// Session 就是打卡要用的 token,但是需要激活才行
	Session    string
	PrivateKey *rsa.PrivateKey
	Key        string
}
