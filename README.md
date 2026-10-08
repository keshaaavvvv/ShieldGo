 ShieldGo

### Layer 7 DoS/DDoS Defense Reverse Proxy

ShieldGo is a Go-based reverse proxy designed to protect HTTP applications from abusive and resource-exhaustion traffic.

It sits between the client and the backend server and applies security controls before forwarding requests.

## Features

- Per-IP rate limiting
- Global rate limiting
- Route-specific rate limiting
- Progressive IP banning
- Request body size limits
- Slow HTTP / Slowloris mitigation
- HTTP request timeouts
- Maximum header-size limits
- Structured JSON security logging

## Architecture

```text
Client
  │
  ▼
ShieldGo :8080
  │
  ├── Rate Limiting
  ├── IP Banning
  ├── Route Protection
  ├── HTTP Timeouts
  └── Security Logging
  │
  ▼
Backend Server
Metasploitable 2 :80
```

## Tech Stack

- **Go**
- **HTTP / Reverse Proxy**
- **Linux**
- **Rate Limiting**
- **Application Security**
- **Network Security**

## Slowloris Protection

ShieldGo uses HTTP timeout controls to protect against slow HTTP requests.

```go
ReadHeaderTimeout: 5 * time.Second,
ReadTimeout:       15 * time.Second,
WriteTimeout:      30 * time.Second,
IdleTimeout:       60 * time.Second,
```

The `ReadHeaderTimeout` limits how long a client can take to send HTTP headers, helping mitigate Slowloris-style attacks.

## Testing

ShieldGo was tested in an isolated **Kali Linux + Metasploitable 2** laboratory.

### Controlled Slowloris Test

- **Target:** ShieldGo `127.0.0.1:8080`
- **Backend:** Metasploitable 2 `192.168.26.129:80`
- **Connections:** 150
- **Header interval:** 10 seconds
- **ReadHeaderTimeout:** 5 seconds

During the test, 150 partially-open connections were established against ShieldGo. After the test, legitimate requests through ShieldGo continued to return:

```text
HTTP/1.1 200 OK
```

## Project Structure

```text
ShieldGo/
├── analysis/
├── internal/
│   ├── ban/
│   ├── config/
│   ├── limiter/
│   └── logging/
├── screenshots/
├── config.example.json
├── go.mod
├── main.go
└── README.md
```

## Run

Clone the repository:

```bash
git clone https://github.com/keshaaavvvv/ShieldGo.git
cd ShieldGo
```

Build:

```bash
go mod tidy
go build -o shieldgo
```

Run:

```bash
./shieldgo -config config.example.json
```

Test the proxy:

```bash
curl -I http://127.0.0.1:8080/
```

## Screenshots

### Slowloris Test

<img width="1375" height="808" alt="Screenshot 2026-10-09 013412" src="https://github.com/user-attachments/assets/8819c074-d641-4e26-997f-56641818343a" />


### Service Availability

<img width="1441" height="813" alt="Screenshot 2026-10-09 013300" src="https://github.com/user-attachments/assets/afb25519-e2fa-4bc6-bbd9-6c1c2611146d" />


### Security Logs
<img width="1403" height="718" alt="Screenshot 2026-10-09 010104" src="https://github.com/user-attachments/assets/82086cc8-3bd0-4960-bebf-fedf9cacb864" />


## Author

**Keshav Chauhan**

B.Tech Cybersecurity | Application Security | Security Engineering

GitHub: [keshaaavvvv](https://github.com/keshaaavvvv)
