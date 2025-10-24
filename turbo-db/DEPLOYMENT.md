# Deployment Guide for Turbo DB

This guide covers deploying Turbo DB to production environments.

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Building for Production](#building-for-production)
4. [Configuration](#configuration)
5. [Deployment Options](#deployment-options)
6. [Monitoring](#monitoring)
7. [Security](#security)
8. [Backup and Recovery](#backup-and-recovery)
9. [Scaling](#scaling)

## Overview

Turbo DB is a database management system that spawns and manages multiple libSQL database instances. In production, you'll want to:

- Run it as a system service
- Enable authentication
- Set up monitoring
- Configure backups
- Implement resource limits

## System Requirements

### Minimum Requirements
- **CPU:** 2 cores
- **RAM:** 2 GB (+ ~50-100MB per database)
- **Disk:** 10 GB (+ storage for databases)
- **OS:** Linux (Ubuntu 20.04+, Debian 11+, RHEL 8+)

### Recommended Requirements
- **CPU:** 4+ cores
- **RAM:** 8+ GB
- **Disk:** SSD with 50+ GB
- **OS:** Ubuntu 22.04 LTS or later

### Software Requirements
- Go 1.20+ (for building)
- systemd (for service management)
- libSQL built with sqld binary

## Building for Production

### 1. Build libSQL

```bash
cd /path/to/libsql
cargo build --release -p libsql-server

# Verify the binary
./target/release/sqld --version
```

### 2. Build Turbo DB

```bash
cd turbo-db

# Build with optimizations
go build -ldflags="-s -w" -o bin/turbo-server ./cmd/server
go build -ldflags="-s -w" -o bin/turbo-cli ./cmd/cli

# Verify
./bin/turbo-server &
./bin/turbo-cli --help
```

### 3. Prepare Installation Directory

```bash
sudo mkdir -p /opt/turbo-db
sudo mkdir -p /var/lib/turbo-db/databases
sudo mkdir -p /var/log/turbo-db
sudo mkdir -p /etc/turbo-db

# Copy binaries
sudo cp bin/turbo-server /opt/turbo-db/
sudo cp bin/turbo-cli /usr/local/bin/
sudo cp ../target/release/sqld /opt/turbo-db/

# Set permissions
sudo chmod +x /opt/turbo-db/turbo-server
sudo chmod +x /opt/turbo-db/sqld
sudo chmod +x /usr/local/bin/turbo-cli
```

## Configuration

### Production Configuration File

Create `/etc/turbo-db/config.env`:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
META_DB_PATH=/var/lib/turbo-db/turbo-meta.db
DATABASES_DIR=/var/lib/turbo-db/databases

# LibSQL Configuration
LIBSQL_BINARY=/opt/turbo-db/sqld
PORT_RANGE_START=9000
PORT_RANGE_END=9500

# Authentication
ENABLE_AUTH=true
API_KEY=CHANGE_THIS_TO_A_SECURE_RANDOM_STRING

# Optional: Logging
LOG_LEVEL=info
```

**Important:** Change the `API_KEY` to a secure random string:
```bash
# Generate a secure API key
openssl rand -hex 32
```

### Environment-Specific Configs

**Development:**
```bash
# /etc/turbo-db/dev.env
SERVER_PORT=8080
ENABLE_AUTH=false
DATABASES_DIR=/var/lib/turbo-db/dev-databases
```

**Staging:**
```bash
# /etc/turbo-db/staging.env
SERVER_PORT=8080
ENABLE_AUTH=true
API_KEY=staging-key-here
DATABASES_DIR=/var/lib/turbo-db/staging-databases
```

**Production:**
```bash
# /etc/turbo-db/production.env
SERVER_PORT=8080
ENABLE_AUTH=true
API_KEY=production-key-here
DATABASES_DIR=/var/lib/turbo-db/databases
PORT_RANGE_START=9000
PORT_RANGE_END=10000
```

## Deployment Options

### Option 1: Systemd Service (Recommended)

Create `/etc/systemd/system/turbo-db.service`:

```ini
[Unit]
Description=Turbo DB Server
After=network.target
Documentation=https://github.com/yourname/turbo-db

[Service]
Type=simple
User=turbo-db
Group=turbo-db
WorkingDirectory=/var/lib/turbo-db
EnvironmentFile=/etc/turbo-db/config.env
ExecStart=/opt/turbo-db/turbo-server
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=turbo-db

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/turbo-db /var/log/turbo-db

# Resource limits
LimitNOFILE=65536
MemoryMax=16G
TasksMax=10000

[Install]
WantedBy=multi-user.target
```

Create the service user:
```bash
sudo useradd -r -s /bin/false turbo-db
sudo chown -R turbo-db:turbo-db /var/lib/turbo-db
sudo chown -R turbo-db:turbo-db /var/log/turbo-db
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable turbo-db
sudo systemctl start turbo-db

# Check status
sudo systemctl status turbo-db

# View logs
sudo journalctl -u turbo-db -f
```

### Option 2: Docker Deployment

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Build Turbo DB
WORKDIR /build
COPY . .
RUN make build

# Runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

# Copy binaries
COPY --from=builder /build/bin/turbo-server /usr/local/bin/
COPY --from=builder /build/bin/turbo-cli /usr/local/bin/

# Copy sqld binary (you need to have this)
COPY target/release/sqld /usr/local/bin/

# Create directories
RUN mkdir -p /var/lib/turbo-db/databases

# Set environment
ENV SERVER_PORT=8080 \
    DATABASES_DIR=/var/lib/turbo-db/databases \
    LIBSQL_BINARY=/usr/local/bin/sqld \
    PORT_RANGE_START=9000 \
    PORT_RANGE_END=9100

EXPOSE 8080 9000-9100

VOLUME ["/var/lib/turbo-db"]

CMD ["/usr/local/bin/turbo-server"]
```

Build and run:
```bash
docker build -t turbo-db:latest .

docker run -d \
  --name turbo-db \
  -p 8080:8080 \
  -p 9000-9100:9000-9100 \
  -v turbo-db-data:/var/lib/turbo-db \
  -e ENABLE_AUTH=true \
  -e API_KEY=your-secure-key \
  turbo-db:latest
```

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  turbo-db:
    image: turbo-db:latest
    container_name: turbo-db
    ports:
      - "8080:8080"
      - "9000-9100:9000-9100"
    volumes:
      - turbo-db-data:/var/lib/turbo-db
      - turbo-db-logs:/var/log/turbo-db
    environment:
      - SERVER_PORT=8080
      - ENABLE_AUTH=true
      - API_KEY=${API_KEY}
      - DATABASES_DIR=/var/lib/turbo-db/databases
      - LIBSQL_BINARY=/usr/local/bin/sqld
      - PORT_RANGE_START=9000
      - PORT_RANGE_END=9100
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  turbo-db-data:
  turbo-db-logs:
```

Run with docker-compose:
```bash
export API_KEY=$(openssl rand -hex 32)
docker-compose up -d
```

### Option 3: Kubernetes Deployment

Create `turbo-db-deployment.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: turbo-db-config
data:
  config.env: |
    SERVER_PORT=8080
    DATABASES_DIR=/var/lib/turbo-db/databases
    LIBSQL_BINARY=/usr/local/bin/sqld
    PORT_RANGE_START=9000
    PORT_RANGE_END=9100
    ENABLE_AUTH=true

---
apiVersion: v1
kind: Secret
metadata:
  name: turbo-db-secret
type: Opaque
stringData:
  api-key: "your-secure-api-key-here"

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: turbo-db-data
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: turbo-db
spec:
  replicas: 1
  selector:
    matchLabels:
      app: turbo-db
  template:
    metadata:
      labels:
        app: turbo-db
    spec:
      containers:
      - name: turbo-db
        image: turbo-db:latest
        ports:
        - containerPort: 8080
          name: api
        - containerPort: 9000
          name: db-start
        env:
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: turbo-db-secret
              key: api-key
        envFrom:
        - configMapRef:
            name: turbo-db-config
        volumeMounts:
        - name: data
          mountPath: /var/lib/turbo-db
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
          limits:
            memory: "16Gi"
            cpu: "4"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: turbo-db-data

---
apiVersion: v1
kind: Service
metadata:
  name: turbo-db
spec:
  selector:
    app: turbo-db
  ports:
  - name: api
    port: 8080
    targetPort: 8080
  type: LoadBalancer
```

Deploy:
```bash
kubectl apply -f turbo-db-deployment.yaml
kubectl get pods
kubectl get svc turbo-db
```

## Monitoring

### Health Checks

```bash
# Check server health
curl http://localhost:8080/health

# List databases
curl -H "X-API-Key: your-key" http://localhost:8080/api/v1/databases
```

### Logging

**View systemd logs:**
```bash
sudo journalctl -u turbo-db -f
```

**View specific database logs:**
```bash
tail -f /var/lib/turbo-db/databases/logs/<database-name>.log
```

### Metrics (Future Enhancement)

You can add Prometheus metrics by extending the API:
- Number of active databases
- Total requests
- Database creation/deletion rate
- Port pool usage
- Memory usage per database

## Security

### 1. Enable Authentication

Always enable authentication in production:
```bash
export ENABLE_AUTH=true
export API_KEY=$(openssl rand -hex 32)
```

### 2. Firewall Configuration

```bash
# Allow API port
sudo ufw allow 8080/tcp

# Allow database port range (only if needed externally)
sudo ufw allow 9000:9100/tcp

# Enable firewall
sudo ufw enable
```

### 3. TLS/SSL

Add a reverse proxy (nginx or Caddy) for TLS:

**Nginx config:**
```nginx
server {
    listen 443 ssl http2;
    server_name turbo-db.example.com;

    ssl_certificate /etc/letsencrypt/live/turbo-db.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/turbo-db.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4. Network Isolation

Run databases on private network:
```bash
# Only bind API to public interface
export SERVER_HOST=0.0.0.0

# Databases will bind to 127.0.0.1 by default
```

## Backup and Recovery

### Backup Strategy

```bash
#!/bin/bash
# /opt/turbo-db/backup.sh

BACKUP_DIR=/var/backups/turbo-db
DATE=$(date +%Y%m%d-%H%M%S)

# Backup metadata database
cp /var/lib/turbo-db/turbo-meta.db $BACKUP_DIR/meta-$DATE.db

# Backup all databases
rsync -av /var/lib/turbo-db/databases/ $BACKUP_DIR/databases-$DATE/

# Clean old backups (keep 7 days)
find $BACKUP_DIR -mtime +7 -delete
```

Add to crontab:
```bash
# Backup every 6 hours
0 */6 * * * /opt/turbo-db/backup.sh
```

### Recovery

```bash
# Stop the service
sudo systemctl stop turbo-db

# Restore metadata
cp /var/backups/turbo-db/meta-20241024-120000.db /var/lib/turbo-db/turbo-meta.db

# Restore databases
rsync -av /var/backups/turbo-db/databases-20241024-120000/ /var/lib/turbo-db/databases/

# Start the service
sudo systemctl start turbo-db
```

## Scaling

### Vertical Scaling

Increase resources for a single instance:
```bash
# Increase port range
export PORT_RANGE_END=20000

# Increase file descriptors
sudo ulimit -n 65536
```

### Horizontal Scaling (Future)

For true horizontal scaling, you'll need to:
1. Implement a database registry (Redis/etcd)
2. Load balancer for API requests
3. Shared storage for databases
4. Database routing layer

## Maintenance

### Updating Turbo DB

```bash
# Download new version
cd /path/to/turbo-db
git pull
make build

# Stop service
sudo systemctl stop turbo-db

# Update binaries
sudo cp bin/turbo-server /opt/turbo-db/

# Start service
sudo systemctl start turbo-db

# Verify
sudo systemctl status turbo-db
```

### Database Cleanup

Regularly clean up stopped/unused databases:
```bash
turbo-cli list | grep stopped | awk '{print $1}' | xargs -I {} turbo-cli delete {} --yes
```

## Troubleshooting

See TESTING.md for detailed troubleshooting steps.

## Production Checklist

- [ ] libSQL binary built and tested
- [ ] Turbo DB binaries built with optimizations
- [ ] Configuration file created with secure API key
- [ ] System user created
- [ ] Directories created with proper permissions
- [ ] Systemd service installed and enabled
- [ ] Firewall configured
- [ ] TLS/SSL configured (via reverse proxy)
- [ ] Monitoring set up
- [ ] Backup strategy implemented
- [ ] Health checks configured
- [ ] Documentation updated
- [ ] Team trained on operations

## Next Steps

After deployment:
1. Monitor logs for errors
2. Set up alerting
3. Configure backups
4. Document runbooks
5. Plan scaling strategy
