# MoviePilot 部署指南

> **版本**: 1.0.0  
> **更新时间**: 2025-12-02

---

## 📋 目录

- [系统要求](#系统要求)
- [部署方式](#部署方式)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [监控配置](#监控配置)
- [备份恢复](#备份恢复)
- [故障排查](#故障排查)

---

## 系统要求

### 最低配置

- **CPU**: 2 核心
- **内存**: 4GB RAM
- **磁盘**: 20GB 可用空间
- **操作系统**: Linux (Ubuntu 20.04+, CentOS 8+)

### 推荐配置

- **CPU**: 4 核心
- **内存**: 8GB RAM
- **磁盘**: 50GB SSD
- **操作系统**: Ubuntu 22.04 LTS

### 软件依赖

- Docker 20.10+
- Docker Compose 2.0+
- Git 2.30+

---

## 部署方式

MoviePilot 支持多种部署方式：

1. **Docker Compose** (推荐) - 适合单机部署
2. **Kubernetes** - 适合集群部署
3. **手动部署** - 适合开发环境

---

## Docker 部署

### 1. 克隆项目

```bash
git clone https://github.com/your-org/moviepilot-go.git
cd moviepilot-go
```

### 2. 配置环境变量

```bash
cd deployments
cp .env.example .env
vim .env
```

**关键配置项**:

```env
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=moviepilot
DB_USER=moviepilot
DB_PASSWORD=your_secure_password

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# 应用配置
GIN_MODE=release
SERVER_PORT=3001
LOG_LEVEL=info

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_ACCESS_EXPIRE=15m
JWT_REFRESH_EXPIRE=7d
```

### 3. 启动服务

**开发环境**:
```bash
make dev
# 或
docker-compose -f deployments/docker-compose.dev.yml up -d
```

**生产环境**:
```bash
make prod
# 或
docker-compose -f deployments/docker-compose.prod.yml up -d
```

### 4. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 查看日志
docker-compose logs -f moviepilot-go

# 健康检查
curl http://localhost:3001/health
```

### 5. 初始化数据库

```bash
# 运行迁移
docker-compose exec moviepilot-go ./main migrate up

# 插入种子数据
docker-compose exec moviepilot-go ./main seed
```

---

## Kubernetes 部署

### 1. 创建命名空间

```bash
kubectl create namespace moviepilot
```

### 2. 创建 Secret

```bash
kubectl create secret generic moviepilot-secrets \
  --from-literal=db-password=your_db_password \
  --from-literal=redis-password=your_redis_password \
  --from-literal=jwt-secret=your_jwt_secret \
  -n moviepilot
```

### 3. 部署 PostgreSQL

```yaml
# postgres-deployment.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: moviepilot
spec:
  serviceName: postgres
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
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: moviepilot
        - name: POSTGRES_USER
          value: moviepilot
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: moviepilot-secrets
              key: db-password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
```

### 4. 部署应用

```yaml
# moviepilot-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: moviepilot-go
  namespace: moviepilot
spec:
  replicas: 3
  selector:
    matchLabels:
      app: moviepilot-go
  template:
    metadata:
      labels:
        app: moviepilot-go
    spec:
      containers:
      - name: moviepilot-go
        image: moviepilot/moviepilot-go:latest
        ports:
        - containerPort: 3001
        env:
        - name: DB_HOST
          value: postgres
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: moviepilot-secrets
              key: db-password
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3001
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 3001
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 5. 创建 Service

```yaml
# moviepilot-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: moviepilot-go
  namespace: moviepilot
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 3001
  selector:
    app: moviepilot-go
```

### 6. 部署

```bash
kubectl apply -f postgres-deployment.yaml
kubectl apply -f moviepilot-deployment.yaml
kubectl apply -f moviepilot-service.yaml
```

---

## 监控配置

### Prometheus + Grafana

**1. 启动监控栈**:

```bash
docker-compose -f deployments/docker-compose.prod.yml up -d prometheus grafana
```

**2. 访问 Grafana**:

- URL: http://localhost:3000
- 默认账号: admin / admin

**3. 导入仪表板**:

- 进入 Grafana
- 导入 `deployments/grafana/dashboards/moviepilot-overview.json`

**4. 配置告警**:

编辑 `deployments/prometheus/alerts.yml` 添加自定义告警规则。

### 关键指标

- **HTTP 请求**: `moviepilot_http_requests_total`
- **API 延迟**: `moviepilot_http_request_duration_seconds`
- **数据库查询**: `moviepilot_db_queries_total`
- **缓存命中率**: `moviepilot_cache_hits_total / (moviepilot_cache_hits_total + moviepilot_cache_misses_total)`
- **活跃订阅**: `moviepilot_subscriptions_active`
- **下载速度**: `moviepilot_download_speed_bytes_per_second`

---

## 备份恢复

### 数据库备份

**手动备份**:
```bash
docker-compose exec postgres pg_dump -U moviepilot moviepilot > backup_$(date +%Y%m%d).sql
```

**自动备份** (cron):
```bash
# 每天凌晨2点备份
0 2 * * * cd /opt/moviepilot && docker-compose exec -T postgres pg_dump -U moviepilot moviepilot > backups/backup_$(date +\%Y\%m\%d).sql
```

### 数据库恢复

```bash
docker-compose exec -T postgres psql -U moviepilot moviepilot < backup_20251202.sql
```

### 配置备份

```bash
# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d).tar.gz configs/ data/
```

---

## 故障排查

### 常见问题

#### 1. 服务无法启动

**检查日志**:
```bash
docker-compose logs moviepilot-go
```

**常见原因**:
- 数据库连接失败
- 端口被占用
- 配置文件错误

#### 2. 数据库连接失败

**检查数据库状态**:
```bash
docker-compose exec postgres pg_isready -U moviepilot
```

**检查网络**:
```bash
docker-compose exec moviepilot-go ping postgres
```

#### 3. 性能问题

**检查资源使用**:
```bash
docker stats
```

**检查慢查询**:
```bash
docker-compose exec postgres psql -U moviepilot -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

#### 4. 内存泄漏

**查看内存使用**:
```bash
curl http://localhost:3001/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### 调试模式

**启用调试日志**:
```bash
# 修改 .env
LOG_LEVEL=debug

# 重启服务
docker-compose restart moviepilot-go
```

### 健康检查

```bash
# 应用健康检查
curl http://localhost:3001/health

# 数据库健康检查
curl http://localhost:3001/health/db

# Redis健康检查
curl http://localhost:3001/health/redis
```

---

## 性能优化

### 1. 数据库优化

```sql
-- 创建索引
CREATE INDEX CONCURRENTLY idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX CONCURRENTLY idx_downloads_status ON downloads(status);

-- 分析表
ANALYZE subscriptions;
ANALYZE downloads;
```

### 2. 缓存优化

```yaml
# 增加 Redis 内存
redis:
  command: redis-server --maxmemory 2gb --maxmemory-policy allkeys-lru
```

### 3. 应用优化

```env
# 增加工作线程
GIN_MODE=release
GOMAXPROCS=4
```

---

## 安全建议

1. **使用强密码**: 数据库、Redis、JWT 密钥
2. **启用 HTTPS**: 使用 Let's Encrypt 证书
3. **限制访问**: 使用防火墙规则
4. **定期更新**: 保持依赖最新
5. **日志审计**: 启用访问日志
6. **备份加密**: 加密备份文件

---

## 扩展部署

### 水平扩展

```bash
# 增加应用实例
docker-compose up -d --scale moviepilot-go=3
```

### 负载均衡

使用 Nginx 作为反向代理：

```nginx
upstream moviepilot {
    server moviepilot-go-1:3001;
    server moviepilot-go-2:3001;
    server moviepilot-go-3:3001;
}

server {
    listen 80;
    server_name moviepilot.example.com;

    location / {
        proxy_pass http://moviepilot;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 联系支持

- **文档**: https://docs.moviepilot.example.com
- **GitHub**: https://github.com/your-org/moviepilot-go
- **Issues**: https://github.com/your-org/moviepilot-go/issues

---

**部署愉快！** 🚀
