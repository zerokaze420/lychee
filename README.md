


<img src="Source/title.png" alt="alt" width="20%">


# LYCHEE


## lychee 荔枝
集成部署CICD

## 已经实现的功能

* 对systemctl 服务的基本监控
* 集成发送飞书

## TODO

1. 实现监测特定服务是否运行正常，记录过滤日志。  
2. 实现集成发送到飞书 / 以及其他交互平台

## 安装

`sudo`运行`./install.sh`  在ubuntu发行版上进行安装


## 构建

```shell
go build -o lychee ./cmd/app/main.go 

```


## 使用

```shell
./lychee -config configs/config.yaml   

```


## 配置文件示例

```shell 

# 监控检查的频率（秒）
check_interval: 60

# systemd 服务监控配置
systemd:
  # 需要监控的服务列表
  services:
    - "sshd.service"
    - "daed.service"

# 飞书机器人通知配置
lark:
  webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_WEBHOOK_ID" # 替换成你的飞书机器人 Webhook 地址

```


