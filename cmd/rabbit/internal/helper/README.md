# Rabbit 服务 JWT 和权限使用说明

## 概述

本文档说明如何在 Rabbit 服务中使用 JWT 中间件和权限获取功能。

## 核心功能

### 1. JWT 中间件 (`JwtServer`)
负责验证和解析 JWT token。

### 2. 团队ID提取中间件 (`ExtractTeamID`)
从 HTTP header 中提取 teamId 并设置到 context 中，支持：
- 从 `X-Team-Id` HTTP header 中获取 teamId
- 自动转换为 uint32 类型并设置到 context

### 3. 权限 Context 操作
提供完整的权限信息 context 管理功能。

## 使用方法

### 在服务器中配置中间件

#### HTTP 服务器配置示例
```go
// 在 cmd/rabbit/internal/server/http.go 中
func NewHTTPServer(bc *conf.Bootstrap, logger log.Logger) *http.Server {
    jwtConf := bc.GetAuth().GetJwt()
    
    // 配置中间件链
    authMiddleware := selector.Server(
        middleware.JwtServer(jwtConf.GetSignKey()),      // JWT验证
        middleware.ExtractTeamID(),                       // 提取teamId
    ).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
    
    opts := []http.ServerOption{
        http.Middleware(
            recovery.Recovery(),
            tracing.Server(),
            metadata.Server(),
            i18n.I18n(),
            logging.Server(logger),
            authMiddleware,    // 应用权限中间件
            middler.Validate(),
        ),
    }
    
    // ... 其他配置
    return http.NewServer(opts...)
}
```

#### gRPC 服务器配置示例
```go
// 在 cmd/rabbit/internal/server/grpc.go 中
func NewGRPCServer(bc *conf.Bootstrap, logger log.Logger) *grpc.Server {
    jwtConf := bc.GetAuth().GetJwt()
    
    // 配置中间件链
    authMiddleware := selector.Server(
        middleware.JwtServer(jwtConf.GetSignKey()),      // JWT验证
        middleware.ExtractTeamID(),                       // 提取teamId
    ).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
    
    opts := []grpc.ServerOption{
        grpc.Middleware(
            recovery.Recovery(),
            tracing.Server(),
            metadata.Server(),
            i18n.I18n(),
            logging.Server(logger),
            authMiddleware,    // 应用权限中间件
            middler.Validate(),
        ),
    }
    
    // ... 其他配置
    return grpc.NewServer(opts...)
}
```

### 在业务代码中使用

#### 在 Service 层获取权限信息
```go
func (s *SendService) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.EmptyReply, error) {
    // 获取团队ID
    teamID, ok := permission.GetTeamIDByContext(ctx)
    if !ok {
        // 没有团队ID，使用默认配置
        log.Info("No team context, using default configuration")
        teamID = 0
    }
    
    // 根据团队ID获取相应的邮件配置
    emailConfig, err := s.getEmailConfigByTeam(ctx, teamID)
    if err != nil {
        return nil, err
    }
    
    // 使用配置发送邮件
    return s.sendEmailWithConfig(ctx, req, emailConfig)
}
```

#### 在 Data 层根据团队获取配置
```go
func (r *configRepo) GetEmailConfig(ctx context.Context) (*EmailConfig, error) {
    // 从context中获取团队ID
    teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
    
    if teamID == 0 {
        // 返回默认配置
        return r.getDefaultEmailConfig(), nil
    }
    
    // 根据团队ID查询配置
    return r.getEmailConfigByTeamID(teamID)
}
```

## HTTP 请求示例

### 发送邮件请求
```bash
curl -X POST http://localhost:8000/v1/send/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "X-Team-Id: 123" \
  -d '{
    "emails": ["user@example.com"],
    "subject": "测试邮件",
    "body": "这是一封测试邮件"
  }'
```

### 发送短信请求
```bash
curl -X POST http://localhost:8000/v1/send/sms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "X-Team-Id: 123" \
  -d '{
    "phones": ["13800138000"],
    "content": "这是一条测试短信"
  }'
```

## Context 工具函数

### 团队ID相关
```go
// 获取团队ID
teamID, ok := permission.GetTeamIDByContext(ctx)

// 获取团队ID，没有时返回0
teamID := permission.GetTeamIDByContextWithZeroValue(ctx)

// 设置团队ID到context
ctx = permission.WithTeamIDContext(ctx, teamID)
```

### 用户ID相关
```go
// 获取用户ID  
userID, ok := permission.GetUserIDByContext(ctx)

// 获取用户ID，没有时返回默认值
userID := permission.GetUserIDByContextWithDefault(ctx, 0)

// 设置用户ID到context
ctx = permission.WithUserIDContext(ctx, userID)
```

### Token相关
```go
// 获取token
token, ok := permission.GetTokenByContext(ctx)

// 设置token到context
ctx = permission.WithTokenContext(ctx, token)
```

### 操作信息相关
```go
// 获取当前操作
operation, ok := permission.GetOperationByContext(ctx)

// 设置操作到context
ctx = permission.WithOperationContext(ctx, operation)
```

## 配置文件

### auth.yaml 配置
```yaml
auth:
  jwt:
    signKey: ${X_MOON_RABBIT_AUTH_JWT_SIGN_KEY:rabbit-sign-key}
    issuer: ${X_MOON_RABBIT_AUTH_JWT_ISSUER:moon.rabbit}
    expire: ${X_MOON_RABBIT_AUTH_JWT_EXPIRE:3600s}
    allowOperations:
      - /api.common.Health/Check
```

## 注意事项

1. **中间件顺序**：
   - 确保 `JwtServer` 在 `ExtractTeamID` 之前
   - 权限相关中间件应该在业务逻辑中间件之前

2. **错误处理**：
   - 如果没有团队ID，业务逻辑应该有降级方案
   - 使用适当的默认值和错误处理

3. **性能考虑**：
   - Context 操作是轻量级的
   - 团队ID解析只在有header时进行

4. **安全性**：
   - 确保JWT token的安全性
   - 验证团队ID的合法性

这个实现提供了简单而强大的权限管理功能，特别适合 Rabbit 通知服务的需求。