# sub2singbox
用于将 clash 或者 Clash.Meta 配置文件，以及订阅链接转换为 sing-box 格式的配置文件。

## 支持协议
- shadowsocks （仅包含 v2ray-plugin, obfs 和 shadow-tls 插件）
- shadowsocksR
- vmess
- vless (含 reality)
- trojan
- socks5
- http
- hysteria
- hysteria2
- tuic5

## 用法

- 首次

  - 准备一个`config-template.json`文件，可以直接使用`https://raw.githubusercontent.com/erguotou520/sub2singbox/refs/heads/main/config-template.json`

  - 准备好你的订阅地址

  - 下载并执行命令

    ```sh
    ./sub2singbox -url https://your.subscription.com/path?token=sometoken|https://your.subscription2.com/path?token=sometoken2
    ```
    命令执行后会在当前目录下生成`sub.json`并根据订阅地址生成`config.json`文件

- 后续
  
  修改当前目录下的`sub.json`，调整相关配置，再次执行`./sub2singbox`