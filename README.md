


<img src="Source/title.png" alt="alt" width="20%">


# LYCHEE


## lychee 荔枝
集成部署CICD

## 已经实现的功能

* 对systemctl 服务的基本监控
* 集成发送飞书
* 基本实现日志检测发送(未测试)
* 实现监测特定服务是否运行正常，记录过滤日志。  

## TODO

1. 实现web界面
2. 实现容器管理


## 安装

`sudo`运行`./install.sh`  在ubuntu发行版上进行安装


## 构建

### 构建要求

* go 1.24.4  

或者使用nix flake 进行构建


```shell
go build -o lychee ./cmd/app/main.go 

```


## 使用

```shell
./lychee -config configs/config.yaml   

```


## 配置文件示例

```yaml
# config.yaml

checkInterval: 60

lark:
  webhookUrl: ""

# systemd 状态监控 (检查服务是否 active)
systemd:
  services:
    - "daed.service"
    - "sshd.service"

# 新增部分：journald 日志监控 (检查服务日志中的关键字)
journal:
  - serviceName: "nginx.service"
    keywords:
      - "error"
      - "failed"
      - "denied"
  - serviceName: "sshd.service"
    keywords:
      - "Failed password"
      - "Invalid user"

```


