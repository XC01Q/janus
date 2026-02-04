# Load Balancing Strategies

## Round Robin

Distributes requests to servers sequentially in a loop.

**When to use:** Homogeneous servers (servers with equal specs), fast requests.

```json
{
  "port": 8080,
  "health_check_time": 5,
  "strategy": "round_robin",
  "backends": [
    {"url": "http://localhost:8081", "weight": 1},
    {"url": "http://localhost:8082", "weight": 1},
    {"url": "http://localhost:8083", "weight": 1}
  ]
}
```

-----

## Weighted

Distribution is proportional to weights — servers with higher weights receive more requests.

**When to use:** Servers of different capacities/power.

```json
{
  "port": 8080,
  "health_check_time": 5,
  "strategy": "weighted",
  "backends": [
    {"url": "http://server1:8080", "weight": 5},
    {"url": "http://server2:8080", "weight": 3},
    {"url": "http://server3:8080", "weight": 2}
  ]
}
```

> With weights 5:3:2, the distribution is: 50%, 30%, 20%

-----

## Least Connections

Directs the request to the server with the minimum number of active connections.

**When to use:** Long-running requests, WebSockets, varying response times.

```json
{
  "port": 8080,
  "health_check_time": 5,
  "strategy": "least_connections",
  "backends": [
    {"url": "http://localhost:8081", "weight": 1},
    {"url": "http://localhost:8082", "weight": 1}
  ]
}
```

-----

## Comparison

| Strategy | Capacity Aware | Load Aware | Best For |
| :--- | :---: | :---: | :--- |
| `round_robin` | ❌ | ❌ | Homogeneous cluster |
| `weighted` | ✅ | ❌ | Diverse servers |
| `least_connections` | ❌ | ✅ | Long requests |
