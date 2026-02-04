# 🛡️ Janus

### Load Balancer written in Go.

Janus is a lightweight load balancer designed for reliability and ease of use. It supports multiple balancing strategies and active health checks.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/XC01Q/janus.git && cd janus

# Run directly
go run ./cmd/proxy
```

## 🛠 Features

* **Balancing Strategies:** Round Robin, Weighted, and Least Connections.
* **Health Checks:** Automatic background monitoring of backend health.
* **Docker Ready:** Containerize and deploy in seconds.
* **Clean Architecture:** Modular design for easy extension.

-----

## ⚙️ Configuration

Create a `config.json` in the root directory:

```json
{
  "port": 8080,
  "strategy": "round_robin",
  "health_check_time": 5,
  "backends": [
    { "url": "http://localhost:8081", "weight": 1 },
    { "url": "http://localhost:8082", "weight": 2 }
  ]
}
```

### Parameters

| Key                 | Default       | Description                                              |
| :------------------ | :------------ | :------------------------------------------------------- |
| `port`              | `8080`        | Proxy listening port.                                    |
| `strategy`          | `round_robin` | Options: `round_robin`, `weighted`, `least_connections`. |
| `health_check_time` | `5`           | Check interval in seconds.                               |

More about balance strategies [there](https://github.com/XC01Q/janus/tree/master/docs/BALANCING_STRATEGIES.md).

-----

## 📦 Deployment

### Using Binary

```bash
go build -o janus ./cmd/proxy
./janus -config config.json
```

### Using Docker

```bash
docker build -t janus .
docker run -p 8080:8080 -v $(pwd)/config.json:/app/config.json janus
```
-----
