package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// 开始打卡
func (u *User) start() (err error) {
	// 获取用户基本信息
	req, err := http.NewRequest("POST", "https://reportedh5.17wanxiao.com/api/clock/school/getUserInfo?appClassify=DK&token="+u.Session, nil)
	if err != nil {
		return err
	}
	// Do it
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	// 解析数据
	resBodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resBlock := struct {
		Code     int    `json:"code"`
		Msg      string `json:"msg"`
		Result   bool   `json:"result"`
		UserInfo struct {
			ClassDescription  string `json:"classDescription"`
			ClassID           int    `json:"classId"`
			CollegeID         int    `json:"collegeId"`
			CustomerAppTypeID int    `json:"customerAppTypeId"`
			CustomerID        int    `json:"customerId"`
			Feature           int    `json:"feature"`
			MajorID           int    `json:"majorId"`
			StuNo             string `json:"stuNo"`
			UserID            int    `json:"userId"`
			Username          string `json:"username"`
		} `json:"userInfo"`
	}{}
	err = json.Unmarshal(resBodyByte, &resBlock)
	if err != nil {
		return err
	}
	u.StdNo = resBlock.UserInfo.StuNo

	// 获取打卡列表
	url := "https://reportedh5.17wanxiao.com/api/clock/school/childApps?appClassify=DK&token=" + u.Session
	req, err = http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	// Do it
	client = &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return nil
	}
	// 解析数据
	resBodyByte, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var list struct {
		AppList []struct {
			ClockPersons            int    `json:"clockPersons"`
			ClockPersonsFlag        bool   `json:"clockPersonsFlag"`
			ClockPersonsType        int    `json:"clockPersonsType"`
			CreateTime              string `json:"createTime"`
			Creator                 string `json:"creator"`
			CustomerAppClassifyID   int    `json:"customerAppClassifyId"`
			CustomerAppTypeRuleList []struct {
				ID int `json:"id"`
			} `json:"customerAppTypeRuleList"`
			DeptName              string  `json:"deptName"`
			EffectiveDistance     int     `json:"effectiveDistance"`
			EffectiveDistanceFlag bool    `json:"effectiveDistanceFlag"`
			ID                    int     `json:"id"`
			ModelFlag             bool    `json:"modelFlag"`
			Modifier              string  `json:"modifier"`
			Name                  string  `json:"name"`
			PeriodEnd             *string `json:"periodEnd,omitempty"`
			PeriodFlag            *bool   `json:"periodFlag,omitempty"`
			PeriodStart           *string `json:"periodStart,omitempty"`
			SchoolClockFlag       bool    `json:"schoolClockFlag"`
			SortNum               int     `json:"sortNum"`
			Status                *bool   `json:"status,omitempty"`
			StatusFlag            *bool   `json:"statusFlag,omitempty"`
			Times                 *int    `json:"times,omitempty"`
			TimesFlag             *bool   `json:"timesFlag,omitempty"`
			URL                   *string `json:"url,omitempty"`
			URLFlag               bool    `json:"urlFlag"`
			UpdateTime            string  `json:"updateTime"`
		} `json:"appList"`
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Result bool   `json:"result"`
	}
	err = json.Unmarshal(resBodyByte, &list)
	if err != nil {
		panic(err)
	}

	for i, app := range list.AppList {
		if app.Name == "校内打卡" && u.InSchoolCheck {
			errmsg := u.inSchoolCheck(list.AppList[i].ID, list.AppList[i].CustomerAppTypeRuleList[0].ID)
			if errmsg != nil {
				logger.Error(errmsg.Error())
				err = fmt.Errorf(errmsg.Error())
			}
		} else if app.Name == "健康打卡" && u.HealthyCheck {
			errmsg := u.healthyCheck()
			if errmsg != nil {
				logger.Error(errmsg.Error())
				err = fmt.Errorf(errmsg.Error())
			}
		}
	}
	return err
}

// 校内打卡
func (u *User) inSchoolCheck(id int, ruleID int) error {
	// 获取打卡参数
	url := "https://reportedh5.17wanxiao.com/api/clock/school/rules?customerAppTypeId=" + strconv.Itoa(id) + "&token=" + u.Session
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	// Do it
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	resBodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// fmt.Println(string(resBodyByte))
	parameterBlock := struct {
		Code               int `json:"code"`
		CustomerAppTypeDto struct {
			CustomerAppClassifyID int `json:"customerAppClassifyId"`
			CustomerCoordinates   []struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"customerCoordinates"`
			EffectiveDistance int    `json:"effectiveDistance"`
			ID                int    `json:"id"`
			Name              string `json:"name"`
			RangeFlag         int    `json:"rangeFlag"`
			RuleList          []struct {
				ClockFlag         bool   `json:"clockFlag"`
				ClockState        int    `json:"clockState"`
				ClockTime         string `json:"clockTime"`
				Clocked           bool   `json:"clocked"`
				CustomerAppTypeID int    `json:"customerAppTypeId"`
				EndTime           string `json:"endTime"`
				ID                int    `json:"id"`
				Latitude          string `json:"latitude"`
				Longitude         string `json:"longitude"`
				Name              string `json:"name"`
				PositionStr       string `json:"positionStr"`
				StartTime         string `json:"startTime"`
				Templateid        string `json:"templateid"`
				TimeFlag          bool   `json:"timeFlag"`
			} `json:"ruleList"`
		} `json:"customerAppTypeDto"`
		Msg    string `json:"msg"`
		Result bool   `json:"result"`
	}{}
	err = json.Unmarshal(resBodyByte, &parameterBlock)
	if err != nil {
		return err
	}

	// 获取上次打卡信息
	jsonMap := make(map[string]interface{})
	jsonMap["businessType"] = "epmpics"
	jsonDataMap := make(map[string]interface{})
	for _, v := range parameterBlock.CustomerAppTypeDto.RuleList {
		if v.ID == ruleID {
			jsonDataMap["templateid"] = v.Templateid
		}
	}
	jsonDataMap["customerAppTypeRuleId"] = ruleID
	jsonDataMap["stuNo"] = u.StdNo
	jsonDataMap["token"] = u.Session
	jsonMap["jsonData"] = jsonDataMap
	jsonMap["method"] = "userComeAppSchool"
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	req, err = http.NewRequest("POST", "https://reportedh5.17wanxiao.com/sass/api/epmpics", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Origin", "https://reportedh5.17wanxiao.com")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "com.eg.android.AlipayGphone")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 5.1.1; vmos Build/LMY48G; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Mobile Safari/537.36 Wanxiao/5.3.4")
	req.Header.Set("x-mass-tappid", "2019030163398604")
	req.Header.Set("Content-Type", "application/json;charset\u003dUTF-8")
	req.Header.Set("Host", "reportedh5.17wanxiao.com")
	req.Header.Set("Connection", "Keep-Alive")
	// Do it
	client = &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	// 解析数据
	resBodyByte, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resblock := struct {
		Code string `json:"code"`
		Data string `json:"data"`
		Msg  string `json:"msg"`
	}{}
	err = json.Unmarshal(resBodyByte, &resblock)
	if err != nil {
		return err
	}
	if resblock.Msg != "成功" {
		return fmt.Errorf("获取上次打卡信息失败")
	}
	lastCheckkData := struct {
		Add                  interface{} `json:"add"`
		AreaStr              string      `json:"areaStr"`
		CardNo               interface{} `json:"cardNo"`
		CusTemplateRelations []struct {
			Assembltype string `json:"assembltype"`
			CheckValues []struct {
				Text string `json:"text"`
			} `json:"checkValues"`
			CreateBy     interface{} `json:"createBy"`
			CreateTime   interface{} `json:"createTime"`
			Decription   string      `json:"decription"`
			DeptName     interface{} `json:"deptName"`
			IsBack       string      `json:"isBack"`
			IsStatistics string      `json:"isStatistics"`
			Params       struct {
			} `json:"params"`
			Placeholder      *string     `json:"placeholder"`
			Propertyname     string      `json:"propertyname"`
			ReadOnly         string      `json:"readOnly"`
			Regex            *string     `json:"regex"`
			Relyon           interface{} `json:"relyon"`
			Relyonvalue      interface{} `json:"relyonvalue"`
			Remark           interface{} `json:"remark"`
			RemarkIsMust     string      `json:"remarkIsMust"`
			Required         string      `json:"required"`
			SearchValue      interface{} `json:"searchValue"`
			ShowValue        interface{} `json:"showValue"`
			SortNum          int         `json:"sortNum"`
			StatisticsValues *string     `json:"statisticsValues"`
			UpdateBy         interface{} `json:"updateBy"`
			UpdateTime       interface{} `json:"updateTime"`
			Value            string      `json:"value"`
		} `json:"cusTemplateRelations"`
		Customerid string `json:"customerid"`
		DeptStr    struct {
			Deptid int    `json:"deptid"`
			Text   string `json:"text"`
		} `json:"deptStr"`
		Isopen       interface{} `json:"isopen"`
		Phonenum     interface{} `json:"phonenum"`
		StuNo        string      `json:"stuNo"`
		SysDeptResVo []struct {
			Childs []struct {
				Childs []struct {
					Childs   []interface{} `json:"childs"`
					DeptName string        `json:"deptName"`
					Deptid   int           `json:"deptid"`
					HasLeaf  bool          `json:"hasLeaf"`
					ParentID int           `json:"parentId"`
				} `json:"childs"`
				DeptName string `json:"deptName"`
				Deptid   int    `json:"deptid"`
				HasLeaf  bool   `json:"hasLeaf"`
				ParentID int    `json:"parentId"`
			} `json:"childs"`
			DeptName string `json:"deptName"`
			Deptid   int    `json:"deptid"`
			HasLeaf  bool   `json:"hasLeaf"`
			ParentID int    `json:"parentId"`
		} `json:"sysDeptResVo"`
		Templateid          string      `json:"templateid"`
		UpDeptStr           interface{} `json:"upDeptStr"`
		UpTime              interface{} `json:"upTime"`
		UpTimeType          interface{} `json:"upTimeType"`
		Updates             interface{} `json:"updates"`
		UserdataDescription interface{} `json:"userdataDescription"`
		Userid              string      `json:"userid"`
		Username            string      `json:"username"`
	}{}
	err = json.Unmarshal([]byte(resblock.Data), &lastCheckkData)
	if err != nil {
		return err
	}

	// 开始打卡
	jsonBlock := struct {
		BusinessType string `json:"businessType"`
		JSONData     struct {
			AreaStr               string `json:"areaStr"`
			ClockState            int    `json:"clockState"`
			CustomerAppTypeRuleID int    `json:"customerAppTypeRuleId"`
			Customerid            int    `json:"customerid"`
			DeptStr               struct {
				Deptid int    `json:"deptid"`
				Text   string `json:"text"`
			} `json:"deptStr"`
			Deptid     int    `json:"deptid"`
			Reportdate int    `json:"reportdate"`
			Source     string `json:"source"`
			StuNo      string `json:"stuNo"`
			Templateid string `json:"templateid"`
			Token      string `json:"token"`
			Updatainfo []struct {
				Propertyname string `json:"propertyname"`
				Value        string `json:"value"`
			} `json:"updatainfo"`
			Userid   int    `json:"userid"`
			Username string `json:"username"`
		} `json:"jsonData"`
		Method string `json:"method"`
		Token  string `json:"token"`
	}{}
	jsonBlock.BusinessType = "epmpics"
	jsonBlock.JSONData.AreaStr = lastCheckkData.AreaStr
	jsonBlock.JSONData.ClockState = 0
	jsonBlock.JSONData.CustomerAppTypeRuleID = ruleID
	customerid, err := strconv.Atoi(lastCheckkData.Customerid)
	if err != nil {
		panic(err)
	}
	jsonBlock.JSONData.Customerid = customerid
	jsonBlock.JSONData.DeptStr = lastCheckkData.DeptStr
	jsonBlock.JSONData.Deptid = lastCheckkData.DeptStr.Deptid
	jsonBlock.JSONData.Reportdate = int(time.Now().UnixNano() / 1e6)
	jsonBlock.JSONData.Source = "app"
	jsonBlock.JSONData.StuNo = lastCheckkData.StuNo
	for _, v := range parameterBlock.CustomerAppTypeDto.RuleList {
		if v.ID == ruleID {
			jsonBlock.JSONData.Templateid = v.Templateid
		}
	}
	jsonBlock.JSONData.Token = u.Session
	for i := 0; i < len(lastCheckkData.CusTemplateRelations); i++ {
		jsonBlock.JSONData.Updatainfo = append(jsonBlock.JSONData.Updatainfo, struct {
			Propertyname string `json:"propertyname"`
			Value        string `json:"value"`
		}{})
		jsonBlock.JSONData.Updatainfo[i].Propertyname = lastCheckkData.CusTemplateRelations[i].Propertyname
		jsonBlock.JSONData.Updatainfo[i].Value = lastCheckkData.CusTemplateRelations[i].Value
	}
	userid, err := strconv.Atoi(lastCheckkData.Userid)
	if err != nil {
		panic(err)
	}
	jsonBlock.JSONData.Userid = userid
	jsonBlock.JSONData.Username = lastCheckkData.Username
	jsonBlock.Method = "submitUpInfoSchool"
	jsonBlock.Token = u.Session
	jsonBytes, err = json.Marshal(jsonBlock)
	if err != nil {
		panic(err)
	}
	req, err = http.NewRequest("POST", "https://reportedh5.17wanxiao.com/sass/api/epmpics", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil
	}
	req.Header.Set("Origin", "https://reportedh5.17wanxiao.com")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "com.eg.android.AlipayGphone")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 5.1.1; vmos Build/LMY48G; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Mobile Safari/537.36 Wanxiao/5.3.4")
	req.Header.Set("x-mass-tappid", "2019030163398604")
	req.Header.Set("Content-Type", "application/json;charset\u003dUTF-8")
	req.Header.Set("Host", "reportedh5.17wanxiao.com")
	req.Header.Set("Connection", "Keep-Alive")
	// Do it
	client = &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return nil
	}
	resBodyByte, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// fmt.Println(string(resBodyByte))
	resBlock := struct {
		Code string `json:"code"`
		Data string `json:"data"`
		Msg  string `json:"msg"`
	}{}
	err = json.Unmarshal(resBodyByte, &resBlock)
	if err != nil {
		return err
	}
	if resBlock.Msg != "成功" {
		return fmt.Errorf("校内打卡失败," + resBlock.Data)
	}
	return nil
}

// 健康打卡
func (u *User) healthyCheck() error {
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
	if err != nil {
		return err
	}

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

	// 解析上一次的打卡信息
	lastHealthyCheck := struct {
		AreaStr              string `json:"areaStr"`
		CusTemplateRelations []struct {
			Propertyname string `json:"propertyname"`
			Value        string `json:"value"`
		} `json:"cusTemplateRelations"`
		Customerid string `json:"customerid"`
		DeptStr    struct {
			Deptid int    `json:"deptid"`
			Text   string `json:"text"`
		} `json:"deptStr"`
		Phonenum   string `json:"phonenum"`
		StuNo      string `json:"stuNo"`
		Templateid string `json:"templateid"`
		Userid     string `json:"userid"`
		Username   string `json:"username"`
	}{}

	err = json.Unmarshal([]byte(data), &lastHealthyCheck)
	if err != nil {
		return err
	}

	healthyCheckData := struct {
		BusinessType string `json:"businessType"`
		JSONData     struct {
			AreaStr    string `json:"areaStr"`
			Customerid string `json:"customerid"`
			DeptStr    struct {
				Deptid int    `json:"deptid"`
				Text   string `json:"text"`
			} `json:"deptStr"`
			Deptid     int    `json:"deptid"`
			GpsType    int    `json:"gpsType"`
			Phonenum   string `json:"phonenum"`
			Reportdate int64  `json:"reportdate"`
			Source     string `json:"source"`
			StuNo      string `json:"stuNo"`
			Templateid string `json:"templateid"`
			Token      string `json:"token"`
			Updatainfo []struct {
				Propertyname string `json:"propertyname"`
				Value        string `json:"value"`
			} `json:"updatainfo"`
			Userid   string `json:"userid"`
			Username string `json:"username"`
		} `json:"jsonData"`
		Method string `json:"method"`
	}{}

	// 填入打卡信息
	healthyCheckData.BusinessType = "epmpics"
	healthyCheckData.Method = "submitUpInfo"
	healthyCheckData.JSONData.AreaStr = lastHealthyCheck.AreaStr
	healthyCheckData.JSONData.Customerid = lastHealthyCheck.Customerid
	healthyCheckData.JSONData.DeptStr = lastHealthyCheck.DeptStr
	healthyCheckData.JSONData.Deptid = lastHealthyCheck.DeptStr.Deptid
	healthyCheckData.JSONData.GpsType = 1
	healthyCheckData.JSONData.Phonenum = ""
	healthyCheckData.JSONData.Reportdate = time.Now().UnixNano() / 1e6
	healthyCheckData.JSONData.Source = "app"
	healthyCheckData.JSONData.StuNo = lastHealthyCheck.StuNo
	healthyCheckData.JSONData.Templateid = lastHealthyCheck.Templateid
	healthyCheckData.JSONData.Token = u.Session
	healthyCheckData.JSONData.Userid = lastHealthyCheck.Userid
	healthyCheckData.JSONData.Username = lastHealthyCheck.Username
	for i := 0; i < len(lastHealthyCheck.CusTemplateRelations); i++ {
		healthyCheckData.JSONData.Updatainfo = append(healthyCheckData.JSONData.Updatainfo, struct {
			Propertyname string "json:\"propertyname\""
			Value        string "json:\"value\""
		}{})
		healthyCheckData.JSONData.Updatainfo[i].Propertyname = lastHealthyCheck.CusTemplateRelations[i].Propertyname
		healthyCheckData.JSONData.Updatainfo[i].Value = lastHealthyCheck.CusTemplateRelations[i].Value
	}
	// 健康打卡
	{
		checkData, err := json.Marshal(healthyCheckData)
		if err != nil {
			return err
		}
		// 初始化一个请求对象
		req, err := http.NewRequest("POST", "https://reportedh5.17wanxiao.com/sass/api/epmpics", bytes.NewBuffer(checkData))
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
		if err != nil {
			return err
		}
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
			return fmt.Errorf("健康打卡失败,%v", data)
		}
		return nil
	}
}

func main() {
	for _, user := range settings.Users {
		// 交换密钥
		err := user.createRSAKey()
		if err != nil {
			logger.Error("创建密钥失败:", err.Error())
			user.report("用户" + user.PhoneNumber + "打卡失败!")
			return
		}
		err = user.exchangeKey()
		if err != nil {
			logger.Error("与服务器交换密钥失败:", err.Error())
			user.report("用户" + user.PhoneNumber + "打卡失败!")

			return
		}
		logger.Info("1. 与服务器交换得到的密钥key: %v和session:%v", user.Key, user.Session)
		// 登录
		err = user.login()
		if err != nil {
			logger.Error("登陆失败:", err.Error())
			user.report("用户" + user.PhoneNumber + "打卡失败!")
			return
		}
		logger.Info("2. 登录成功！")
		// 激活Session
		err = user.activateSession()
		if err != nil {
			logger.Error("Session激活失败, %v", err)
			user.report("用户" + user.PhoneNumber + "打卡失败!")
			return
		}
		logger.Info("3. Session激活为Token成功！")
		// 打卡
		err = user.start()
		if err != nil {
			user.report("用户" + user.PhoneNumber + "打卡失败:如果你设置了两种打卡那至少有一种失败了")
			return
		}
		logger.Info("4. 打卡成功！")
		// 发送通知
		user.report("用户" + user.PhoneNumber + "打卡成功!")
	}
}
