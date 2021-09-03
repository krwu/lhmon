# 腾讯云轻量应用服务器流量检测工具

本工具可用于自动检测指定帐号和区域下的腾讯云轻量应用服务器的流量包使用情况。
根据设置，当某个轻量应用服务器的流量包使用率达到一定值时，可以告警、自动关闭轻量应用服务器。

## 使用方法：

- 配置文件
  首先要准备一个 yaml 配置文件，格式如下： 

  ```yaml
  sct_key: SERVER_SEND_KEY # Server酱(sct.ftqq.com)的 SendKey
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
