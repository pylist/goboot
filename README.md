# Goboot

一个基于 Go 的现代化 REST API 后端框架，提供完整的用户认证和管理功能。

## 特性

- **用户认证** - 注册、登录、JWT 双令牌机制（Access Token + Refresh Token）
- **权限控制** - 基于角色的访问控制（RBAC），区分普通用户和管理员
- **用户管理** - 完整的用户 CRUD、密码重置、状态管理
- **安全性** - bcrypt 密码加密、令牌黑名单、软删除
- **日志系统** - 结构化日志输出，支持文件轮转

## 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | Fiber |
| ORM | GORM |
| 数据库 | MySQL |
| 缓存 | Redis |
| 认证 | JWT |
| 配置管理 | Viper |

## 项目结构

```
goboot/
├── main.go                 # 入口文件
├── config.yaml             # 配置文件
├── config/                 # 配置管理
├── internal/               # 内部代码
│   ├── handler/            # HTTP 处理器
│   ├── service/            # 业务逻辑层
│   ├── model/              # 数据模型
│   ├── middleware/         # 中间件
│   └── repository/         # 数据访问层
├── pkg/                    # 公共包
│   ├── database/           # 数据库连接
│   ├── logger/             # 日志工具
│   ├── utils/              # 工具函数
│   └── response/           # 响应封装
└── router/                 # 路由定义
```

## 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+

### 安装

```bash
git clone https://github.com/your-username/goboot.git
cd goboot
go mod download
```

### 配置

创建 `config.yaml` 文件：

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  mode: "debug"  # debug, release, test

mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "your_password"
  dbname: "goboot"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100

jwt:
  access_secret: "your_access_secret"
  refresh_secret: "your_refresh_secret"
  access_expire: 2      # 小时
  refresh_expire: 168   # 小时（7天）

log:
  level: "debug"
  file_path: "./logs/app.log"
  max_size: 100         # MB
  max_backups: 3
  max_age: 28           # 天
```

### 运行

```bash
go run main.go
```

服务将启动在 `http://127.0.0.1:8080`

## API 文档

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| POST | `/api/auth/refreshToken` | 刷新令牌 |
| POST | `/api/auth/logout` | 退出登录 |

### 用户接口（需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/profile` | 获取个人信息 |
| POST | `/api/user/updateProfile` | 更新个人信息 |
| POST | `/api/user/changePassword` | 修改密码 |

### 管理员接口（需管理员权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/user/list` | 用户列表（分页） |
| POST | `/api/admin/user/add` | 创建用户 |
| GET | `/api/admin/user/detail` | 用户详情 |
| POST | `/api/admin/user/update` | 更新用户 |
| POST | `/api/admin/user/delete` | 删除用户 |
| POST | `/api/admin/user/resetPassword` | 重置密码 |
| POST | `/api/admin/user/updateStatus` | 更新状态 |

### 请求示例

**登录：**
```bash
curl -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "123456"}'
```

**访问受保护接口：**
```bash
curl http://127.0.0.1:8080/api/user/profile \
  -H "Authorization: Bearer <access_token>"
```

### 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 依赖说明

| 依赖 | 用途 |
|------|------|
| gin-gonic/gin | Web 框架 |
| gorm.io/gorm | ORM 框架 |
| golang-jwt/jwt | JWT 令牌 |
| redis/go-redis | Redis 客户端 |
| spf13/viper | 配置管理 |
| golang.org/x/crypto | 密码加密 (bcrypt) |
| natefinch/lumberjack | 日志轮转 |
| shopspring/decimal | 精确十进制运算 |

## License

MIT
