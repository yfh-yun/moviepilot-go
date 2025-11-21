# 部署指南

## 🚀 部署概览

MoviePilot Go 支持多种部署方式，从本地开发到生产环境的完整部署方案。

## 📋 环境要求

### 最低配置
- **CPU**: 2 核心
- **内存**: 4GB RAM
- **存储**: 20GB 可用空间
- **操作系统**: Linux (Ubuntu 20.04+, CentOS 8+) / macOS / Windows 10+

### 推荐配置
- **CPU**: 4 核心
- **内存**: 8GB RAM
- **存储**: 100GB SSD
- **网络**: 1Gbps 带宽

### 依赖服务
- **Go**: 1.21+ (开发环境)
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **PostgreSQL**: 14+
- **Redis**: 6+

## 🐳 Docker 部署（推荐）

### 1. 快速启动

```bash
# 克隆项目
git clone https://github.com/yfh-yun/moviepilot-go.git
cd moviepilot-go

# 启动所有服务
docker-compose -f deployments/docker-compose.yml up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

### 2. 生产环境配置

创建生产环境配置文件：
```bash
cp deployments/docker-compose.prod.yml docker-compose.yml
```

编辑环境变量：
```bash
# .env
POSTGRES_DB=moviepilot
POSTGRES_USER=moviepilot
POSTGRES_PASSWORD=your-secure-password
REDIS_PASSWORD=your-redis-password
JWT_SECRET=your-jwt-secret-key
```

启动生产环境：
```bash
docker-compose up -d
```

### 3. 服务配置详解

#### docker-compose.yml
```yaml
version: '3.8'

services:
  app:
    image: moviepilot-go:latest
    container_name: moviepilot-app
    restart: unless-stopped
    ports:
      - "3001:3001"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=moviepilot
      - DB_USER=moviepilot
      - DB_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./data:/app/data
      - ./configs:/app/configs
    depends_on:
      - postgres
      - redis
    networks:
      - moviepilot-network

  postgres:
    image: postgres:14-alpine
    container_name: moviepilot-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - moviepilot-network

  redis:
    image: redis:6-alpine
    container_name: moviepilot-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - moviepilot-network

  plugins:
    image: moviepilot-plugins:latest
    container_name: moviepilot-plugins
    restart: unless-stopped
    ports:
      - "5000:5000"
    environment:
      - GRPC_SERVER_HOST=0.0.0.0
      - GRPC_SERVER_PORT=5000
    volumes:
      - ./plugins:/app/plugins
    networks:
      - moviepilot-network

  nginx:
    image: nginx:alpine
    container_name: moviepilot-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - app
    networks:
      - moviepilot-network

volumes:
  postgres_data:
  redis_data:

networks:
  moviepilot-network:
    driver: bridge
```

## 🔧 本地开发部署

### 1. 环境准备

```bash
# 安装 Go (使用 gvm 推荐)
curl -sSL https://git.io/g-install | sh
gvm install go1.21 -B
gvm use go1.21 --default

# 安装依赖
go mod download

# 安装 PostgreSQL
sudo apt-get install postgresql postgresql-contrib

# 安装 Redis
sudo apt-get install redis-server
```

### 2. 数据库设置

```bash
# 创建数据库
sudo -u postgres createdb moviepilot

# 创建用户
sudo -u postgres createuser moviepilot

# 设置密码
sudo -u postgres psql
ALTER USER moviepilot PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE moviepilot TO moviepilot;
\q
```

### 3. 配置文件

```bash
# 复制配置文件
cp configs/config.yaml.sample configs/config.yaml

# 编辑配置
vim configs/config.yaml
```

配置示例：
```yaml
server:
  host: "0.0.0.0"
  port: 3001
  mode: "debug"  # debug, release

database:
  host: "localhost"
  port: 5432
  name: "moviepilot"
  user: "moviepilot"
  password: "your-password"
  ssl_mode: "disable"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret"
  expires_in: 24h

logging:
  level: "info"
  format: "json"
  output: "stdout"

plugins:
  enabled: true
  grpc_address: "localhost:5000"
```

### 4. 启动应用

```bash
# 启动主应用
go run cmd/server/main.go

# 启动插件服务 (新终端)
cd python-plugins
python cmd/server/main.py
```

## ☁️ 云平台部署

### Kubernetes 部署

#### 1. 命名空间和配置
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: moviepilot
---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: moviepilot-config
  namespace: moviepilot
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 3001
    database:
      host: "postgres-service"
      port: 5432
      name: "moviepilot"
      user: "moviepilot"
    redis:
      host: "redis-service"
      port: 6379
```

#### 2. 数据库部署
```yaml
# postgres-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: moviepilot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        env:
        - name: POSTGRES_DB
          value: "moviepilot"
        - name: POSTGRES_USER
          value: "moviepilot"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: moviepilot-secrets
              key: postgres-password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: moviepilot
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

#### 3. 应用部署
```yaml
# app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: moviepilot-app
  namespace: moviepilot
spec:
  replicas: 3
  selector:
    matchLabels:
      app: moviepilot-app
  template:
    metadata:
      labels:
        app: moviepilot-app
    spec:
      containers:
      - name: app
        image: moviepilot-go:latest
        ports:
        - containerPort: 3001
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
      volumes:
      - name: config-volume
        configMap:
          name: moviepilot-config
---
apiVersion: v1
kind: Service
metadata:
  name: moviepilot-service
  namespace: moviepilot
spec:
  selector:
    app: moviepilot-app
  ports:
  - port: 80
    targetPort: 3001
  type: LoadBalancer
```

### AWS ECS 部署

#### 1. 任务定义
```json
{
  "family": "moviepilot",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "moviepilot-app",
      "image": "your-account.dkr.ecr.region.amazonaws.com/moviepilot-go:latest",
      "portMappings": [
        {
          "containerPort": 3001,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_HOST",
          "value": "your-rds-endpoint"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/moviepilot",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

## 🔒 安全配置

### 1. SSL/TLS 配置

#### Nginx 配置
```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;

    location / {
        proxy_pass http://app:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}
```

### 2. 防火墙配置
```bash
# UFW 配置
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 3. 密钥管理
```bash
# 生成 JWT 密钥
openssl rand -base64 32

# 生成数据库密码
openssl rand -base64 16

# 使用 Kubernetes Secrets
kubectl create secret generic moviepilot-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=postgres-password=$(openssl rand -base64 16) \
  --from-literal=redis-password=$(openssl rand -base64 16)
```

## 📊 监控和日志

### 1. Prometheus 配置
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'moviepilot'
    static_configs:
      - targets: ['app:3001']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### 2. Grafana 仪表板
- 系统性能监控
- API 请求统计
- 数据库性能指标
- 错误率和响应时间

### 3. 日志聚合
```yaml
# filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
    - add_docker_metadata:
        host: "unix:///var/run/docker.sock"

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "moviepilot-%{+yyyy.MM.dd}"
```

## 🔄 备份和恢复

### 1. 数据库备份
```bash
# 自动备份脚本
#!/bin/bash
BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)

docker exec postgres pg_dump -U moviepilot moviepilot > $BACKUP_DIR/backup_$DATE.sql

# 保留最近 7 天的备份
find $BACKUP_DIR -name "backup_*.sql" -mtime +7 -delete
```

### 2. 配置备份
```bash
# 备份配置文件
tar -czf configs_backup_$(date +%Y%m%d).tar.gz configs/

# 备份插件数据
tar -czf plugins_backup_$(date +%Y%m%d).tar.gz plugins/
```

## 🚨 故障排查

### 1. 常见问题

#### 应用无法启动
```bash
# 检查日志
docker-compose logs app

# 检查端口占用
netstat -tulpn | grep 3001

# 检查配置文件
docker-compose config
```

#### 数据库连接失败
```bash
# 测试数据库连接
docker exec postgres psql -U moviepilot -d moviepilot -c "SELECT 1;"

# 检查网络连接
docker exec app ping postgres
```

### 2. 性能调优

#### 数据库优化
```sql
-- 创建索引
CREATE INDEX idx_media_title ON media(title);
CREATE INDEX idx_media_type ON media(type);
CREATE INDEX idx_transfers_status ON transfers(status);

-- 分析查询性能
EXPLAIN ANALYZE SELECT * FROM media WHERE title LIKE '%keyword%';
```

#### 应用优化
```bash
# 调整 Go 运行时参数
export GOMAXPROCS=4
export GOGC=100

# 调整连接池大小
# configs/config.yaml
database:
  max_idle_conns: 20
  max_open_conns: 200
  conn_max_lifetime: "1h"
```

---

**注意**: 生产环境部署前请务必进行充分测试，并制定详细的运维预案。