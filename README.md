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

## 参数验证器

项目内置了参数验证器 `pkg/validator`，支持结构体标签验证，自动返回中文错误信息。

### 基本用法

```go
import "goboot/pkg/validator"

// 定义请求结构体
type LoginRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50" label:"用户名"`
    Password string `json:"password" validate:"required,min=6" label:"密码"`
    Email    string `json:"email" validate:"email" label:"邮箱"`
}

// 在 Handler 中使用
func (h *Handler) Login(c fiber.Ctx) error {
    var req LoginRequest
    if err := validator.BindAndValidate(c, &req); err != nil {
        return err  // 自动返回标准错误响应
    }
    // 验证通过，继续处理...
}
```

### 验证规则

| 规则 | 说明 | 示例 |
|------|------|------|
| `required` | 必填字段 | `validate:"required"` |
| `min` | 最小长度（字符串）或最小值（数字） | `validate:"min=3"` |
| `max` | 最大长度（字符串）或最大值（数字） | `validate:"max=50"` |
| `len` | 精确长度 | `validate:"len=11"` |
| `range` | 长度/值范围 | `validate:"range=3-50"` |
| `email` | 邮箱格式 | `validate:"email"` |
| `phone` | 中国手机号 | `validate:"phone"` |
| `url` | URL 格式 | `validate:"url"` |
| `ip` | IP 地址 | `validate:"ip"` |
| `alpha` | 纯字母 | `validate:"alpha"` |
| `alphanum` | 字母和数字 | `validate:"alphanum"` |
| `numeric` | 纯数字字符串 | `validate:"numeric"` |
| `number` | 数字（含负数和小数） | `validate:"number"` |
| `lowercase` | 小写字母 | `validate:"lowercase"` |
| `uppercase` | 大写字母 | `validate:"uppercase"` |
| `username` | 用户名（字母、数字、下划线） | `validate:"username"` |
| `password` | 密码强度（必须包含字母和数字） | `validate:"password=6"` |
| `idcard` | 中国身份证号 | `validate:"idcard"` |
| `contains` | 包含指定字符串 | `validate:"contains=@"` |
| `startswith` | 以指定字符串开头 | `validate:"startswith=http"` |
| `endswith` | 以指定字符串结尾 | `validate:"endswith=.com"` |
| `oneof` | 枚举值（空格分隔） | `validate:"oneof=male female"` |
| `eq` | 等于 | `validate:"eq=10"` |
| `ne` | 不等于 | `validate:"ne=0"` |
| `gt` | 大于 | `validate:"gt=0"` |
| `gte` | 大于等于 | `validate:"gte=1"` |
| `lt` | 小于 | `validate:"lt=100"` |
| `lte` | 小于等于 | `validate:"lte=99"` |
| `regex` | 正则表达式 | `validate:"regex=^[a-z]+$"` |

### 组合规则

多个规则用逗号分隔：

```go
type User struct {
    Username string `validate:"required,min=3,max=20,username" label:"用户名"`
    Age      int    `validate:"required,gte=0,lte=150" label:"年龄"`
    Role     string `validate:"required,oneof=admin user guest" label:"角色"`
}
```

### 错误消息

验证器会根据 `label` 标签自动生成中文错误消息：

- `用户名不能为空`
- `密码长度不能小于6`
- `邮箱必须是有效的邮箱地址`
- `手机号必须是有效的手机号`
- `年龄必须大于或等于0`

### 自定义错误消息

```go
// 全局设置
validator.SetMessage("required", "{field}是必填项")
validator.SetMessage("min", "{field}最少需要{param}个字符")
```

### 自定义验证规则

```go
// 注册自定义验证器
validator.RegisterValidator("even", func(field reflect.Value, param string) bool {
    if field.Kind() == reflect.Int {
        return field.Int()%2 == 0
    }
    return false
})

// 使用自定义规则
type Request struct {
    Number int `validate:"even" label:"数字"`
}
```

### Fiber 集成方法

| 方法 | 说明 |
|------|------|
| `validator.BindAndValidate(c, &req)` | 绑定 Body 并验证 |
| `validator.BindQueryAndValidate(c, &req)` | 绑定 Query 参数并验证 |
| `validator.MustValidate(c, &req)` | 仅验证（不绑定） |
| `validator.Validate(&req)` | 直接验证结构体 |

## 依赖说明

| 依赖 | 用途 |
|------|------|
| gofiber/fiber | Web 框架 |
| gorm.io/gorm | ORM 框架 |
| golang-jwt/jwt | JWT 令牌 |
| redis/go-redis | Redis 客户端 |
| spf13/viper | 配置管理 |
| golang.org/x/crypto | 密码加密 (bcrypt) |
| natefinch/lumberjack | 日志轮转 |

## License

MIT
