# 部署文档

## 概述

MoviePilot Go 支持多种部署方式，包括Docker、Kubernetes和传统部署。

## 系统要求

### 最低配置
- CPU: 2核心
- 内存: 4GB
- 存储: 20GB
- 网络: 100Mbps

### 推荐配置
- CPU: 4核心
- 内存: 8GB
- 存储: 100GB SSD
- 网络: 1Gbps

### 软件依赖
- Docker 20.10+
- Docker Compose 2.0+
- PostgreSQL 14+
- Redis 6+

## Docker部署

### 快速启动

1. 克隆项目
```bash
git clone https://github.com/moviepilot/moviepilot-go.git
cd moviepilot-go
```

2. 配置环境
```bash
cp configs/config.yaml.sample configs/config.yaml
# 编辑配置文件
```

3. 启动服务
```bash
docker-compose -f deployments/docker-compose.yml up -d
```

4. 验证部署
```bash
curl http://localhost:3001/health
```

### 环境配置

#### 开发环境
```bash
docker-compose -f deployments/docker-compose.dev.yml up -d
```

包含以下服务：
- PostgreSQL
- Redis
- pgAdmin (数据库管理)
- Redis Commander (Redis管理)

#### 生产环境
```bash
# 设置环境变量
export VERSION=latest
export DB_PASSWORD=your-secure-password
export REDIS_PASSWORD=your-secure-password
export JWT_SECRET=your-jwt-secret
export API_KEY=your-api-key

# 启动生产环境
docker-compose -f deployments/docker-compose.prod.yml up -d
```

### 配置文件

#### docker-compose.yml
完整的服务编排配置，包含：
- 应用服务
- 数据库服务
- 缓存服务
- 监控服务

#### docker-compose.prod.yml
生产环境优化配置：
- Nginx反向代理
- 日志轮转
- 资源限制
- 安全配置

## Kubernetes部署

### 准备工作

1. 创建命名空间
```bash
kubectl create namespace moviepilot
```

2. 创建ConfigMap
```bash
kubectl apply -f deployments/k8s/configmap.yaml
```

3. 创建Secret
```bash
kubectl apply -f deployments/k8s/secret.yaml
```

### 部署应用

```bash
kubectl apply -f deployments/k8s/
```

### 服务清单

- `deployment.yaml`: 应用部署配置
- `service.yaml`: 服务暴露配置
- `ingress.yaml`: Ingress配置
- `configmap.yaml`: 配置文件
- `secret.yaml`: 敏感信息
- `pvc.yaml`: 持久卷声明

### 监控

```bash
# 查看Pod状态
kubectl get pods -n moviepilot

# 查看服务状态
kubectl get services -n moviepilot

# 查看日志
kubectl logs -f deployment/moviepilot-go -n moviepilot
```

## 传统部署

### 二进制部署

1. 下载二进制文件
```bash
wget https://github.com/moviepilot/moviepilot-go/releases/latest/download/moviepilot-go-linux-amd64.tar.gz
tar -xzf moviepilot-go-linux-amd64.tar.gz
```

2. 安装依赖
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql redis-server

# CentOS/RHEL
sudo yum install postgresql-server redis
```

3. 配置数据库
```bash
sudo -u postgres createdb moviepilot
sudo -u postgres createuser moviepilot
sudo -u postgres psql -c "ALTER USER moviepilot PASSWORD 'moviepilot123';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE moviepilot TO moviepilot;"
```

4. 配置应用
```bash
cp configs/config.yaml.sample configs/config.yaml
# 编辑配置文件
```

5. 启动应用
```bash
./moviepilot-go
```

### 源码编译部署

1. 安装Go环境
```bash
# 下载并安装Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

2. 克隆并编译
```bash
git clone https://github.com/moviepilot/moviepilot-go.git
cd moviepilot-go
go mod download
go build -o moviepilot-go cmd/server/main.go
```

3. 部署步骤同二进制部署

## 系统服务配置

### Systemd服务

创建服务文件 `/etc/systemd/system/moviepilot-go.service`:

```ini
[Unit]
Description=MoviePilot Go
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=moviepilot
Group=moviepilot
WorkingDirectory=/opt/moviepilot-go
ExecStart=/opt/moviepilot-go/moviepilot-go
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable moviepilot-go
sudo systemctl start moviepilot-go
```

## 负载均衡

### Nginx配置

```nginx
upstream moviepilot_go {
    server 127.0.0.1:3001;
    server 127.0.0.1:3002;
    server 127.0.0.1:3003;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://moviepilot_go;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 数据库优化

### PostgreSQL优化

配置文件 `postgresql.conf`:

```ini
# 内存配置
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# 连接配置
max_connections = 100
listen_addresses = '*'

# 日志配置
log_statement = 'all'
log_duration = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '

# 性能配置
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

### Redis优化

配置文件 `redis.conf`:

```ini
# 内存配置
maxmemory 512mb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000

# 网络配置
tcp-keepalive 300
timeout 0

# 安全配置
requirepass your-redis-password
```

## 监控和日志

### Prometheus监控

1. 配置Prometheus抓取应用指标
2. 配置Grafana仪表板
3. 设置告警规则

### 日志管理

1. 配置日志轮转
2. 集中化日志收集
3. 日志分析和告警

## 备份策略

### 数据库备份

```bash
# 每日备份
pg_dump -h localhost -U moviepilot moviepilot | gzip > backup_$(date +%Y%m%d).sql.gz

# 自动备份脚本
#!/bin/bash
BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U moviepilot moviepilot | gzip > $BACKUP_DIR/moviepilot_$DATE.sql.gz
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

### 配置备份

```bash
# 备份配置文件
tar -czf configs_backup_$(date +%Y%m%d).tar.gz configs/
```

## 安全配置

### 防火墙配置

```bash
# 开放必要端口
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

### SSL证书

```bash
# 使用Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## 故障排除

### 常见问题

1. **应用无法启动**
   - 检查配置文件
   - 检查数据库连接
   - 查看应用日志

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数
   - 检查网络连通性

3. **性能问题**
   - 检查系统资源
   - 分析慢查询
   - 优化数据库配置

### 日志分析

```bash
# 查看应用日志
tail -f logs/app.log

# 查看数据库日志
tail -f /var/log/postgresql/postgresql-14-main.log

# 查看系统日志
journalctl -u moviepilot-go -f
```

## 升级指南

### 版本升级

1. 备份数据
2. 停止服务
3. 更新代码
4. 运行数据库迁移
5. 重启服务
6. 验证功能

### 回滚操作

1. 停止服务
2. 恢复代码
3. 恢复数据库
4. 重启服务
5. 验证功能