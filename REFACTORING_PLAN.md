# MoviePilot-Go 架构重构计划

## 重构目标
解决架构分析中发现的问题，提升代码质量和可维护性。

## 重构原则
- 渐进式重构，避免大规模重写
- 保持向后兼容性
- 优先解决高影响问题
- 每步都要保证系统可运行

## 阶段一：统一命名规范（高优先级）

### 1.1 目录重命名计划
```
internal/api/ → internal/apis/
internal/service/ → internal/services/
internal/model/ → internal/models/
internal/repository/ → internal/repositories/
internal/actions/ → internal/workflows/  # 更符合语义
internal/scheduler/ → internal/schedulers/
```

### 1.2 文件重命名计划
```
handlers/ → handlers/  # 保持复数
middleware/ → middlewares/
validator/ → validators/
```

## 阶段二：合并冗余模块（高优先级）

### 2.1 合并service和modules
```
internal/services/ + internal/modules/ → internal/business/
├── domains/          # 领域模型
├── services/         # 业务服务
├── workflows/        # 工作流（原actions）
└── policies/         # 业务策略
```

### 2.2 合并foundation和core
```
internal/foundation/ + internal/core/ → internal/infrastructure/
├── config/           # 配置管理
├── security/         # 安全组件
├── events/           # 事件系统
└── context/          # 上下文管理
```

## 阶段三：优化配置管理（中优先级）

### 3.1 统一配置结构
```
config/
├── core/             # 核心配置
├── environments/     # 环境配置
├── validation/       # 配置验证
└── providers/        # 配置提供者
```

## 实施步骤

### Step 1: 备份当前状态
### Step 2: 重命名目录（保持向后兼容）
### Step 3: 更新导入路径
### Step 4: 合并冗余模块
### Step 5: 更新文档
### Step 6: 测试验证

## 风险控制
- 每个步骤都要提交代码
- 保持测试通过
- 更新相关文档
- 通知团队成员