# 腾讯云轻量应用服务器流量检测工具

本工具可用于自动检测指定帐号和区域下的腾讯云轻量应用服务器的流量包使用情况。
根据设置，当某个轻量应用服务器的流量包使用率达到一定值时，可以告警、自动关闭轻量应用服务器。

## 使用方法：

- API帐号
  1. 在腾讯云"[访问管理](https://console.cloud.tencent.com/cam/overview)-用户-[用户列表](https://console.cloud.tencent.com/cam)" 下面，创建一个新的子用户（或者使用一个现有的子用户），该子用户的访问方式应该为__“编程访问”__，不需要控制台访问权限。
  2. 选定的子用户最小所需权限如下：
    - lighthouse:DescribeInstances 用于读取轻量云实例信息
    - lighthouse:DescribeInstancesTrafficPackages 用于读取轻量云流量包信息
    - lighthouse:StopInstances 用于自动关机（如果不需要自动关机功能，可以不授予此项权限，见下面的配置说明）

- 配置文件
  首先要准备一个 yaml 配置文件，格式如下： 

  ```yaml
  notify_method: nextrt # 使用 nextrt 渠道进行通知
  nextrt_type: email # NextRT 的推送使用 Email 方式
  nextrt_token: 3413d...0ad1 # NextRT 的推送 Token（到 https://app.nextrt.com 注册）
  warn_rate: 0.75 # 报警通知的流量使用率，如果设为 0 表示不使用报警功能
  shutdown_rate: 0.9 # 自动关机的流量使用率，如果设为 0 表示不使用自动关机功能
  check_interval: 30 # 检查间隔，单位：秒
  accounts: # 要检查的账户列表
    - name: "账户一" # 账户名称
      secret_id: "secret_id_1" # 该帐户的 secretId
      secret_key: "secret_key_1" # 该账户的 secretKey
      regions: [ "ap-hongkong", "ap-guangzhou" ] # 要监控的区域
    - name: "账户二" # 账户名称
      secret_id: "secret_id_2" # 该帐户的 secretId
      secret_key: "secret_key_2" # 该账户的 secretKey
      regions: [ "ap-guangzhou" ] # 要监控的区域
  ```
  **配置文件特别说明：**
  1. `notify_method` 目前可选的值有：sct(Server酱)、werobot（企业微信群机器人）、NextRT，三选一
  2. 根据 `notify_method` 的不同，需要配置做对应的配置：
     - sct: ([https://sct.ftqq.com/](https://sct.ftqq.com/r/13200))
       ```yaml
       notify_method: sct
       sct_key: SCT63835...RhMSG # 从 sct.ftqq.com 获取的 sendkey
       ```
     - werobot:
       ```yaml
       notify_method: werobot
       werobot_webhook: https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=077...5f4 # 企业微信群机器人的 webhook 地址
       werobot_chatid:  # 企业微信群机器人推送通知的 chatid，没有可留空
       ```
     - nextrt:
       ```yaml
       notify_method: nextrt
       nextrt_type: Telegram # NextRT 的推送类型：Email,Telegram,WorkWechat,Lark,Dingding,Android
       nextrt_token: 3413...0ad1 # NextRT 的推送 Token，从 https://app.nextrt.com/ 注册获取
       ```
- 启动 Docker 容器： 

  ```bash
  docker run -itd --name lhmon -v ${yaml配置文件路径}:/etc/lhmon/conf.yml -v /etc/localtime:/etc/localtime kairee/lhmon:latest
  ```
  如果担心日志文件大小，可以用 `--log-opt max-size=5m --log-opt max-file=3` 来指定（__注意：仅限默认未配置 docker 日志参数的情况，如果你已全局配置，或者日志驱动不是 `json-file`，请根据自己的情况具体配置__）。
- 或者使用 docker-compose： 

  ```yaml
  version: "3"
  services:
    lhmon:
      image: kairee/lhmon:latest
      restart: unless-stopped
      volumes:
        - /etc/localtime:/etc/localtime
        - ${yaml配置文件路径}:/etc/lhmon/conf.yml
      logging:
        driver: "json-file" # 默认的日志驱动
        options:
          max-size: "5m" # 单个日志文件最大尺寸
          max-file: "5" # 最多保留日志文件数量
  ```
- 以上命令或配置中的 `${yaml配置文件路径}` 请自行替换为**自己的路径**

## 开发计划：

- [ ] 支持企业微信机器人直接推送通知到企业微信
- [ ] 提供 web 界面进行管理配置和查看流量使用历史记录
