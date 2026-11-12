# Cloudflare Tunnel 公网发布与 ChatOps 免密直登指南

> 本文档指导如何借助 Cloudflare Tunnel 将本地 `tmux-ui` 服务安全穿透至公网，并配合 URL 免密参数实现在 Telegram Mini App、飞书、Discord 等应用内的无感秒登。

---

## 1. 核心优势

- **无需公网 IP 与路由器端口映射**：即便处于大内网（CGNAT）或移动热点后，也能畅通穿透。
- **自带免维护全球 HTTPS / TLS 证书**：自动防范中间人劫持与明文窥视。
- **原生支持全双工 WebSocket**：`tmux-ui` 的实时 ANSI 终端流与状态广播完全支持穿透。
- **安全零暴露**：真实服务器 IP 完全隐藏在 Cloudflare Anycast 边缘网络后。

---

## 2. 模式一：Quick Tunnel（快速临时隧道，推荐尝鲜体验）

若尚未配置专属域名，可直接使用 Cloudflare 提供的免登录临时隧道：

### 步骤 1：安装 `cloudflared` CLI
- **macOS (Homebrew)**: `brew install cloudflare/cloudflare/cloudflared`
- **Linux (Debian/Ubuntu)**:
  ```bash
  curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
  sudo dpkg -i cloudflared.deb
  ```
- **Windows**: `winget install --id Cloudflare.cloudflared`

### 步骤 2：启动 tmux-ui 并指定安全 Token
```bash
./bin/tmux-ui --port 8080 --token my_super_secret_token_123456
```

### 步骤 3：一键暴露
```bash
cloudflared tunnel --url http://127.0.0.1:8080
```
控制台会输出一个类似如下的临时公共域名：
```
https://random-words-1234.trycloudflare.com
```

### 步骤 4：生成 ChatOps 免密秒登链接
在公网链接末尾附带你的 Token 参数：
```
https://random-words-1234.trycloudflare.com/?token=my_super_secret_token_123456
```
将此链接直接发到 Telegram、飞书或 Discord，点击即可无感秒登并进入实时监控控制台！

---

## 3. 模式二：Named Tunnel（生产级固定域名，推荐长期运行）

### 步骤 1：登录与创建专属隧道
```bash
cloudflared tunnel login
cloudflared tunnel create tmux-ui-tunnel
```
记录返回的 Tunnel ID（例如 `3d2984b6-xxxx-xxxx-xxxx-xxxxxxxxxxxx`）。

### 步骤 2：配置 DNS 路由
将你的专属子域名（如 `tmux.yourdomain.com`）绑定到该隧道：
```bash
cloudflared tunnel route dns tmux-ui-tunnel tmux.yourdomain.com
```

### 步骤 3：编写配置文件
参考项目中的 `config/tunnel.yml.example`，创建 `/etc/cloudflared/config.yml`：
```yaml
tunnel: 3d2984b6-xxxx-xxxx-xxxx-xxxxxxxxxxxx
credentials-file: /root/.cloudflared/3d2984b6-xxxx-xxxx-xxxx-xxxxxxxxxxxx.json

ingress:
  - hostname: tmux.yourdomain.com
    service: http://127.0.0.1:8080
    originRequest:
      connectTimeout: 30s
      noTLSVerify: false
  - service: http_status:404
```

### 步骤 4：安装并注册为系统服务（Systemd）
```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
```

### 步骤 5：启动 tmux-ui
```bash
./bin/tmux-ui --port 8080 --token YOUR_TOKEN --public-url https://tmux.yourdomain.com
```
此时系统控制台将高亮输出：
```
Public Access URL: https://tmux.yourdomain.com/?token=YOUR_TOKEN
```

---

## 4. Telegram Mini App / 飞书无感免密直登

在 Telegram Bot 中配置 Web App 按钮或在飞书群机器人中发送卡片消息时，将 URL 设置为：
```
https://tmux.yourdomain.com/?token=<YOUR_TOKEN>
```
用户点击后：
1. **自动提取**：页面加载时自动从 URL 中提取 `token`。
2. **会话持久化**：Token 暂存至会话存储（`sessionStorage`），多标签或刷新不会丢失。
3. **地址栏净化**：前端无刷新移除 URL 上的 `?token=...`，防止用户在分享当前网页时泄漏凭证。
4. **全链路注入**：WebSocket 和所有后端 REST API 自动附带 Token，秒级渲染终端。
