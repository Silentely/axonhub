# 解决方案总结：自定义模型端点支持

## 问题描述

用户在使用 Cloudflare AI Gateway 等自定义 Base URI 时，自动获取模型功能会失败，返回 401 Unauthorized 错误。错误示例：

```
获取模型失败:failed to fetch models:GET- https://gateway.ai.cloudflare.com/v1/f29be4114f2232323232/cosr/v1/models with status 401 Unauthorized
```

## 问题根源

Cloudflare AI Gateway 的 URL 结构为：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/{provider}/...
```

当用户配置的 Base URL 包含代理/网关路径时，系统自动添加的 `/v1/models` 可能会导致 URL 结构不正确，因为缺少了提供商特定的路径段。

## 解决方案

### 1. 数据库层面

在 `channels` 表中添加新字段 `custom_models_endpoint`（可选，可为 NULL）：

**文件**: `internal/ent/schema/channel.go`
```go
field.String("custom_models_endpoint").
    Optional().Nillable().
    Comment("Custom endpoint for fetching models, used for proxies like Cloudflare AI Gateway"),
```

### 2. 业务逻辑层面

更新模型获取逻辑，支持自定义端点：

**文件**: `internal/server/biz/model_fetcher.go`

- 在 `FetchModelsInput` 结构体中添加 `CustomModelsEndpoint *string` 字段
- 在 `FetchModels` 方法中，优先使用自定义端点：
  ```go
  if input.CustomModelsEndpoint != nil && *input.CustomModelsEndpoint != "" {
      modelsURL = strings.TrimSuffix(*input.CustomModelsEndpoint, "/")
      authHeaders = make(http.Header)
      // Set Anthropic headers if needed
  } else {
      modelsURL, authHeaders = f.prepareModelsEndpoint(channelType, input.BaseURL)
  }
  ```

### 3. 自动同步支持

**文件**: `internal/server/biz/channel_model_sync.go`

更新自动同步逻辑，传递 `CustomModelsEndpoint` 参数：
```go
result, err := modelFetcher.FetchModels(ctx, FetchModelsInput{
    ChannelType:          ch.Type.String(),
    BaseURL:              ch.BaseURL,
    ChannelID:            lo.ToPtr(ch.ID),
    CustomModelsEndpoint: ch.CustomModelsEndpoint,
})
```

### 4. GraphQL API 支持

GraphQL schema 自动通过 Ent 生成更新，支持在创建和更新渠道时设置 `customModelsEndpoint`。

### 5. 前端支持

**文件**: `frontend/src/features/channels/data/schema.ts`

在 Channel schema 中添加字段：
```typescript
customModelsEndpoint: z.string().optional().nullable(),
```

## 使用方法

### 方法 1: 完整的 Base URL（推荐）

在创建或更新渠道时，在 Base URL 中包含完整路径：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai
```

系统会自动添加 `/v1/models`，生成正确的 URL：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models
```

### 方法 2: 使用自定义端点

如果需要更精确的控制，可以通过以下方式设置：

#### GraphQL API
```graphql
mutation {
  updateChannel(id: "channel-id", input: {
    customModelsEndpoint: "https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models"
  }) {
    id
    customModelsEndpoint
  }
}
```

#### 直接数据库更新
```sql
UPDATE channels 
SET custom_models_endpoint = 'https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models'
WHERE id = channel_id;
```

## 兼容性

- ✅ 向后兼容：新字段为可选，不影响现有渠道
- ✅ 自动迁移：Ent 会自动处理数据库迁移
- ✅ 所有渠道类型：适用于 OpenAI、Anthropic、Gemini 等所有渠道类型
- ✅ 自动同步：与 `auto_sync_supported_models` 功能完全兼容

## 工作流程

1. **配置渠道**：设置 Base URL 或 Custom Models Endpoint
2. **启用自动同步**（可选）：启用 `auto_sync_supported_models`
3. **自动获取**：系统每小时自动获取模型列表
4. **手动获取**：也可在管理界面手动点击"获取模型"按钮
5. **认证处理**：系统自动添加正确的认证头（API Key、Anthropic-Version 等）

## 受影响的文件

### 后端
- `internal/ent/schema/channel.go` - 添加字段定义
- `internal/server/biz/model_fetcher.go` - 更新获取逻辑
- `internal/server/biz/channel_model_sync.go` - 更新自动同步
- `internal/ent/*` - Ent 生成的代码
- `internal/server/gql/*` - GraphQL 生成的代码

### 前端
- `frontend/src/features/channels/data/schema.ts` - 添加类型定义

### 文档
- `docs/custom-models-endpoint.md` - 详细使用文档

## 测试建议

1. **正常渠道测试**：确保未设置自定义端点的渠道仍能正常工作
2. **Cloudflare 测试**：使用 Cloudflare AI Gateway 配置测试
3. **自动同步测试**：验证定时任务是否正确使用自定义端点
4. **手动获取测试**：在管理界面测试"获取模型"功能
5. **不同提供商测试**：测试 OpenAI、Anthropic、Gemini 等不同类型

## 未来改进

- [ ] 在前端 UI 中添加自定义端点配置表单
- [ ] 添加端点有效性验证
- [ ] 提供常见网关的预设模板
- [ ] 添加端点测试功能
