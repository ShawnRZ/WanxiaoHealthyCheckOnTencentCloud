# WanxiaoHealthyCheckOnTencentCloud

------

>>>特别说明：本项目理论上试用于 ___各地大多数学校___ 的完美校园健康打卡（注意不是校内打卡），但是仅测试过 ___河南师范大学___，烦请测试能用的同学留下学校信息以便后人查看，佛系更新 _（这个项目本就是为了偷懒而存在的）_ ，有问题可以提[Issues][2]。

------

## 功能

- 完美校园每日定时自动打卡

- pushplus推送，[官网][11]

- 邮箱推送

- server酱推送（有问题），[官网][10]

> 三种推送方式请任选其一（直接把对应的key填入配置文件即可），推荐使用邮箱或者pushplus，若填写了多种推送方式则按照优先级选择最大的，其优先级为：邮箱>pushplus>server酱

## 寻人启事

急需熟悉Android开发或者逆向的童鞋协助开发，QQ：1713252605

## 实现原理

- 模拟手机app登录，获取打卡所需的token

- 获取上次打卡的信息，并用于自动打卡

- 由于完美校园APP对登录设备的IMEI做了检测，如果是新设备只能使用短信验证码登录。而且，在高版本的Android上似乎对IMEI做了加密处理。所以，强烈建议用安装低版本Android系统的模拟器上先登录一次，再将模拟器的IMEI拷贝用作本项目的模拟登录，具体见使用方法。

## 使用方法（图片显示不了，可去博客，博客地址在最下方）：

1. 下载所需文件

     - [VMOS模拟器][3]，若链接失效可去[VMOS官网][4]下载

     - [完美校园APP][5]，若链接失效可去[完美校园官网][6]下载

2. 打开VMOS，并安装完美校园APP进行打卡

    - 打开VMOS，点击“+”![VMOS打开页面](http://blog.rzx.ink/usr/uploads/2021/03/2089543289.jpg "VMOS打开页面")

    - 选择安卓5.1极客版下载并添加![添加虚拟机页面](http://blog.rzx.ink/usr/uploads/2021/03/162672475.jpg "添加虚拟机页面")

    - 等待一会后，会自动进入虚拟机，点击右侧小圆点
    ![虚拟机页面](http://blog.rzx.ink/usr/uploads/2021/03/3491046275.jpg "虚拟机页面")

    - 选择文件传输
    ![悬浮窗](http://blog.rzx.ink/usr/uploads/2021/03/2518043956.jpg "悬浮窗")

    - 选择“我要导入”
    ![文件导入](http://blog.rzx.ink/usr/uploads/2021/03/171272120.jpg "选择“我要导入”")

    - 导入完美校园安装包

    - 等待一会后，返回虚拟机，打开完美校园进行打卡

3. 获取虚拟机IMEI

    - 进入设置，注意不要选错了图标
    ![进入设置](http://blog.rzx.ink/usr/uploads/2021/03/4023518919.jpg "进入设置")

    - 下滑，找到“关于手机”
    ![](http://blog.rzx.ink/usr/uploads/2021/03/1079664733.jpg)

    - 点击“状态信息”
    ![](http://blog.rzx.ink/usr/uploads/2021/03/3198369272.jpg)

    - 点击“IMEI信息”
    ![](http://blog.rzx.ink/usr/uploads/2021/03/476623448.jpg)

    - 将IMEI信息记录备用

4. 点击右上角绿色按钮“Code”，选择“Download ZIP”[下载][7]压缩包

5. 解压压缩包会得到一个文件夹，用记事本打开文件夹中的users.json文本，按提示输入。

6. 打开[腾讯云][8]，在右上角点击登录，登陆后进入控制台，在搜索栏中搜索“云函数”

7. 进入云函数后，选择左侧边栏的函数服务，再点击“新建”，设置如图，点击上传将刚刚下载的文件夹
![](http://blog.rzx.ink/usr/uploads/2021/03/1386830612.png "云函数设置")

8. 配置触发器，图中设置为每天1点触发，可通过修改Cron表达式更改时间，具体见[文档][9]
![](http://blog.rzx.ink/usr/uploads/2021/03/2247204310.png)

9. 最后点击完成，部署成功！

## 个人博客地址：[RZX's blog][1]

[1]:http://blog.rzx.ink
[2]:https://github.com/FNDHSTD/WanxiaoHealthyCheckOnTencentCloud/issues
[3]:https://files.vmos.cn/vmospro/version/2021012018500427995_vmoscn.apk
[4]:http://www.vmos.cn/
[5]:http://apk.17wanxiao.com/campus/apk/wanxiao.apk
[6]:https://www.17wanxiao.com/new/index.html
[7]:https://github.com/FNDHSTD/WanxiaoHealthyCheckOnTencentCloud/archive/master.zip
[8]:https://cloud.tencent.com/
[9]:https://cloud.tencent.com/document/product/583/9708
[10]:http://sc.ftqq.com/
[11]:https://pushplus.hxtrip.com/