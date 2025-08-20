
<img src="Source/title.png" alt="alt" width="20%">

# LYCHEE: A Powerful Tool for Automated Operations & Intelligent Monitoring 🚀

**LYCHEE** is a **command-line tool** that integrates **CI/CD deployment**, **system monitoring**, and **alert notifications**. It is designed to simplify your operational workflow, ensure the healthy operation of your system services, and send timely notifications when issues arise. Say goodbye to tedious tasks and embrace efficiency! ✨

> Warning: This project is currently in dev mode. There may be significant changes in the future. Issues are welcome!

-----

## Core Features 💡

- [x] **Systemd Service Monitoring:** Provides basic and effective monitoring for `systemctl` services to ensure they are running correctly. 👁️‍🗨️
- [x] **Lark Integration:** Seamlessly sends alerts and notifications to your Lark groups. 📨
- [x] **Basic Log Anomaly Detection:** Monitors service logs for specific keywords to help you detect potential issues early (currently a basic implementation, pending comprehensive testing). 🔍
- [x] **Service Health Checks:** Actively checks if specified services are running correctly, and records and filters relevant logs for analysis. ❤️‍🩹
- [x] **Multi-Account Log Forwarding:** Enhanced log forwarding feature that supports sending logs to multiple accounts or destinations. 📧
- [ ] **Container Management:** Support for monitoring and managing containerized applications. 🐳


-----

## Installation 🛠️

Installing LYCHEE on **Ubuntu distributions** is straightforward.

Simply run the installation script with `sudo`:

```bash
sudo ./install.sh
````

-----

## Build from Source 🏗️

### Build Requirements

  * **Go 1.24.4** or higher

You can also use **Nix Flake** for a reproducible build environment.

To build the executable, run:

```bash
go build -o lychee ./cmd/app/main.go
```

-----

### Build from Nix

```shell
nix build .
```

## Usage 🚀

After building or installing, you can run LYCHEE by specifying a configuration file:

```bash
./lychee -config configs/config.yaml
```

-----

## Configuration File Example ⚙️

Here is a sample `config.yaml` to help you get started:

```yaml
# config.yaml

# The frequency in seconds at which LYCHEE checks service status and logs. ⏱️
checkInterval: 60

# Lark bot Webhook URL for sending notifications. 🔔
lark:
  WebhookURLs:
    - "[https://open.feishu.cn/open-apis/bot/v2/hook/URLA](https://open.feishu.cn/open-apis/bot/v2/hook/URLA)"
    - "[https://open.feishu.cn/open-apis/bot/v2/hook/URLB](https://open.feishu.cn/open-apis/bot/v2/hook/URLB)"

# --- Systemd Service Monitoring ---
# A list of systemd services to monitor. LYCHEE will check if they are in an 'active' state. ✅
systemd:
  services:
    - "daed.service"
    - "sshd.service"
    - "nginx.service"

# --- Journald Log Monitoring ---
# Configure log monitoring for specific services and keywords.
# LYCHEE will send an alert if any of the specified keywords are found in the service's Journal logs. 🚨
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

## Contributing 🤝

We welcome contributions\! Please see our [Contributing Guide](https://www.google.com/search?q=CONTRIBUTING.md) for more information.

## License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

