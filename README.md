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
    - 若启用 SSL 证书清理，另需：
      - ssl:DescribeCertificates
      - ssl:CreateCertificateBindResourceSyncTask
      - ssl:DescribeCertificateBindResourceTaskResult
      - ssl:DeleteCertificate（仅告警不删时可省略）

- 配置文件
  首先要准备一个 yaml 配置文件，格式如下： 

  ```yaml
  notify_method: sct # 使用 sct 渠道进行通知
  sct_key: SCT63835...RhMSG # 从 sct.ftqq.com 获取的 sendkey
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
  1. `notify_method` 目前可选的值有：sct(Server酱)、werobot（企业微信群机器人）、notifyx（NotifyX）、telegram（Telegram机器人），四选一
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
     - notifyx: ([https://www.notifyx.cn](https://www.notifyx.cn))
       ```yaml
       notify_method: notifyx
       notifyx_key: YOUR_NOTIFYX_KEY # NotifyX 的推送 Key（必填，到 https://www.notifyx.cn 注册获取）
       notifyx_team: YOUR_TEAM       # NotifyX 推送目标团队（可选，留空则推送到默认团队）
       ```
     - telegram:
       ```yaml
       notify_method: telegram
       telegram_bot_token: 123456:ABC-DEF... # Telegram 机器人的 Bot Token（通过 @BotFather 创建）
       telegram_user_id: "987654321"         # 接收通知的 Telegram 用户 ID（可通过 @userinfobot 获取）
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


- SSL 证书自动清理（可选）

  在配置中增加 `ssl` 段即可（与流量监控共用 `accounts` 密钥与通知渠道）：

  ```yaml
  ssl:
    enabled: false          # 总开关
    auto_delete: true       # 无关联则删除
    dry_run: true           # 默认 true：模拟删除，确认无误后再改为 false
    # expire_days: 30       # 可选；省略则用官方「即将过期」过滤（约 30 天）
    check_interval: 86400   # 默认每天一次
  ```

  行为简述：
  1. 找出**已过期**（状态码 3）以及**即将过期**的证书；
  2. 通过腾讯云关联资源异步查询确认是否绑定 CLB/CDN/WAF 等资源；
  3. **仅当无关联**且 `auto_delete=true`、`dry_run=false` 时调用删除；
  4. 查询失败或超时时**不删除**，只告警。

  实现上使用腾讯云 Go SDK 的 Common Client（只依赖 `tencentcloud/common`，不引入 `tencentcloud/ssl` 产品包）。

## 开发计划：

- [x] SSL 证书过期/即将过期清理（无关联才删；Common Client）
- [ ] 支持企业微信机器人直接推送通知到企业微信
- [ ] 提供 web 界面进行管理配置和查看流量使用历史记录
- [ ] 将 lighthouse 调用迁移到 Common Client，去掉产品包依赖
