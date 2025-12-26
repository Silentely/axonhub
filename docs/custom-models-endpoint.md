# 自定义模型端点

## 概述

当使用 AI 网关或代理服务（例如 Cloudflare AI Gateway）时，标准的模型列表 API 端点可能无法正常工作。`custom_models_endpoint` 字段允许你为每个渠道指定一个自定义的模型列表获取端点。

## 问题场景

### Cloudflare AI Gateway 示例

当使用 Cloudflare AI Gateway 时，URL 结构如下：

```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/{provider}/...
```

例如，对于 OpenAI：
- 正常 OpenAI API: `https://api.openai.com/v1/models`
- 通过 Cloudflare: `https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models`

### 常见错误

如果你在 Base URL 中配置了不完整的 Cloudflare 网关 URL：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}
```

系统会自动添加 `/v1/models`，导致URL变成：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/v1/models
```

这是错误的，因为缺少了提供商路径部分（`/openai`、`/anthropic` 等）。

## 解决方案

### 方法 1: 完整的 Base URL（推荐）

在 Base URL 字段中包含完整的提供商路径：

```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai
```

这样系统会正确添加 `/v1/models`，最终 URL 为：
```
https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models
```

### 方法 2: 使用自定义模型端点

如果你需要更精确的控制，可以使用 `custom_models_endpoint` 字段：

#### 通过 GraphQL API

```graphql
mutation {
  updateChannel(id: "your-channel-id", input: {
    customModelsEndpoint: "https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models"
  }) {
    id
    customModelsEndpoint
  }
}
```

#### 通过数据库直接更新

```sql
UPDATE channels 
SET custom_models_endpoint = 'https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/openai/v1/models'
WHERE id = your_channel_id;
```

## 工作原理

当设置了 `custom_models_endpoint` 时：

1. **自动获取模型**：当渠道启用了 `auto_sync_supported_models` 时，系统会使用自定义端点定期获取模型列表
2. **手动获取模型**：在管理界面中点击"获取模型"按钮时，也会使用自定义端点
3. **认证**：系统会自动添加正确的认证头（API Key、Anthropic-Version 等）

## 适用场景

以下情况可能需要使用自定义模型端点：

1. **Cloudflare AI Gateway**：如上所述
2. **企业代理**：企业内部的 API 网关可能有自定义的端点结构
3. **自定义 API 服务器**：如果你使用了自己的 API 聚合服务
4. **特殊提供商**：某些 AI 提供商可能有非标准的 API 端点结构

## 注意事项

1. `custom_models_endpoint` 是可选的，大多数标准提供商不需要设置
2. 当设置了自定义端点时，系统会完全忽略自动端点构建逻辑
3. 确保自定义端点 URL 是完整的，包括所有必要的路径参数
4. 认证头会根据渠道类型自动添加，无需在 URL 中包含 API Key

## 测试

设置自定义端点后，你可以：

1. 在管理界面中点击"获取模型"按钮测试
2. 检查错误日志以查看详细的请求信息
3. 使用浏览器开发工具查看实际的 API 请求

## 示例配置

### Cloudflare + OpenAI

```json
{
  "type": "openai",
  "baseURL": "https://gateway.ai.cloudflare.com/v1/abc123/my-gateway/openai",
  "customModelsEndpoint": null  // 不需要，baseURL 已经完整
}
```

或者：

```json
{
  "type": "openai",
  "baseURL": "https://gateway.ai.cloudflare.com/v1/abc123/my-gateway/openai",
  "customModelsEndpoint": "https://gateway.ai.cloudflare.com/v1/abc123/my-gateway/openai/v1/models"
}
```

### Cloudflare + Anthropic

```json
{
  "type": "anthropic",
  "baseURL": "https://gateway.ai.cloudflare.com/v1/abc123/my-gateway/anthropic",
  "customModelsEndpoint": "https://gateway.ai.cloudflare.com/v1/abc123/my-gateway/anthropic/v1/models"
}
```

## 相关字段

- `base_url`: 渠道的基础 URL，用于所有 API 请求
- `custom_models_endpoint`: 自定义的模型列表端点（可选）
- `auto_sync_supported_models`: 是否自动同步支持的模型列表
- `supported_models`: 渠道支持的模型列表
