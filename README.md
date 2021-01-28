# WanxiaoHealthyCheckOnTencentCloud

------

>>>特别说明：本项目只试用于 ___河南师范大学___ 已经 ___离校___ 学生的 ___完美校园___ 健康打卡，佛系更新 _（这个项目本就是为了偷懒而存在的）_ ，确有其他需求可以提[Issues][2]。

------

## 实现原理

- 模拟手机app登录，获取打卡所需的token

- 获取上次打卡的信息，并用于自动打卡

- 由于完美校园APP对登录设备的IMEI做了检测，如果是新设备只能使用短信验证码登录。而且，在高版本的Android上似乎对IMEI做了加密处理。所以，强烈建议用安装低版本Android系统的模拟器上先登录一次，再将模拟器的IMEI拷贝用作本项目的模拟登录，具体见使用方法。

## 使用方法：

1. 下载所需文件

     - [VMOS模拟器][3]，若链接失效可去[VMOS官网][4]下载

     - [完美校园APP][5]，若链接失效可去[完美校园官网][6]下载

2. 打开VMOS，并安装完美校园APP进行打卡

    - 打开VMOS，点击“+”![VMOS打开页面](http://blog.rzx.ink/usr/uploads/2021/01/1660309503.jpg "VMOS打开页面")

    - 选择安卓5.1极客版下载并添加![添加虚拟机页面](http://blog.rzx.ink/usr/uploads/2021/01/1808473139.png "添加虚拟机页面")

    - 等待一会后，会自动进入虚拟机，点击右侧小圆点
    ![虚拟机页面](http://blog.rzx.ink/usr/uploads/2021/01/4284627275.png "虚拟机页面")

    - 选择文件传输
    ![悬浮窗](http://blog.rzx.ink/usr/uploads/2021/01/3883076510.jpg "悬浮窗")

    - 选择“我要导入”
    ![文件导入](http://blog.rzx.ink/usr/uploads/2021/01/3347114332.jpg "选择“我要导入”")

    - 导入完美校园安装包
    ![导入完美校园安装包](http://blog.rzx.ink/usr/uploads/2021/01/4138725774.png "导入完美校园安装包")

    - 等待一会后，返回虚拟机，打开完美校园进行打卡

3. 获取虚拟机IMEI

    - 进入设置，注意不要选错了图标
    ![进入设置](http://blog.rzx.ink/usr/uploads/2021/01/3508598477.png "进入设置")

    - 下滑，找到“关于手机”
    ![](http://blog.rzx.ink/usr/uploads/2021/01/3502767582.png)

    - 点击“状态信息”
    ![](http://blog.rzx.ink/usr/uploads/2021/01/4090614745.png)

    - 点击“IMEI信息”
    ![](http://blog.rzx.ink/usr/uploads/2021/01/2538324620.png)

    - 将IMEI信息记录备用



## 个人博客地址：[RZX's blog][1]

[1]:http://blog.rzx.ink
[2]:https://github.com/FNDHSTD/WanxiaoHealthyCheckOnTencentCloud/issues
[3]:https://files.vmos.cn/vmospro/version/2021012018500427995_vmoscn.apk
[4]:http://www.vmos.cn/
[5]:http://apk.17wanxiao.com/campus/apk/wanxiao.apk
[6]:https://www.17wanxiao.com/new/index.html
