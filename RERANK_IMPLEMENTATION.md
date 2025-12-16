# Rerank (重排序) 功能实现文档

本文档描述了在 AxonHub 项目中实现的 Rerank 功能。

## 概述

Rerank 功能允许用户通过 API 对文档列表进行重新排序，根据查询的相关性为每个文档打分。

## 实现架构

### 1. 数据结构层 (Data Layer)

文件：`internal/objects/rerank.go`

定义了三个核心数据结构：

- **RerankRequest**: 包含 `model` (string), `query` (string), `documents` ([]string), `top_n` (int, 可选)
- **RerankResponse**: 包含 `results` ([]RerankResult) 和 `usage` (*Usage)
- **RerankResult**: 包含 `index` (int), `relevance_score` (float64), `document` (string, 可选)
- **Usage**: 包含 token 使用量信息

### 2. Transformer 层

#### 接口定义 (interfaces.go)

文件：`internal/llm/transformer/interfaces.go`

添加了新的 `Transformer` 接口，扩展了 `Outbound` 接口并添加了 `Rerank` 方法：

```go
type Transformer interface {
    Outbound
    Rerank(ctx context.Context, req *objects.RerankRequest) (*objects.RerankResponse, error)
}
```

#### OpenAI 实现

文件：`internal/llm/transformer/openai/rerank.go`

实现了 `OutboundTransformer` 的 `Rerank` 方法，支持：
- 标准 OpenAI API 格式
- Azure OpenAI API 格式
- 自动构建正确的 URL 和认证头
- 错误处理和响应解析

### 3. 业务逻辑层 (Biz Layer)

文件：`internal/server/biz/rerank.go`

实现了 `RerankService`，提供以下功能：

- 验证请求参数
- 查找支持指定模型的 Channel
- 调用 Transformer 的 Rerank 方法
- 记录日志和错误处理

依赖注入已在 `fx_module.go` 中配置。

### 4. API 层

文件：`internal/server/api/rerank.go`

实现了 `RerankHandlers`，提供 HTTP 端点：

- 解析 JSON 请求体
- 调用业务逻辑层
- 返回 JSON 响应

依赖注入已在 `fx_module.go` 中配置。

### 5. 路由注册

文件：`internal/server/routes.go`

在 OpenAI 兼容的 API 组中注册了新的路由：

```
POST /v1/rerank
```

该路由受以下中间件保护：
- API Key 认证
- 超时控制
- 链路追踪

### 6. 前端集成

文件：`frontend/src/lib/api-client.ts`

添加了 `rerankApi` 对象，提供客户端方法：

```typescript
rerankApi.rerank({
  model: string,
  query: string,
  documents: string[],
  top_n?: number
})
```

## 使用示例

### cURL 请求

```bash
curl -X POST http://localhost:8090/v1/rerank \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-3-small",
    "query": "What is machine learning?",
    "documents": [
      "Machine learning is a subset of artificial intelligence.",
      "The weather is nice today.",
      "Deep learning uses neural networks."
    ],
    "top_n": 2
  }'
```

### 响应示例

```json
{
  "results": [
    {
      "index": 0,
      "relevance_score": 0.95,
      "document": "Machine learning is a subset of artificial intelligence."
    },
    {
      "index": 2,
      "relevance_score": 0.78,
      "document": "Deep learning uses neural networks."
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 0,
    "total_tokens": 25
  }
}
```

### TypeScript 客户端使用

```typescript
import { rerankApi } from '@/lib/api-client'

const result = await rerankApi.rerank({
  model: 'text-embedding-3-small',
  query: 'What is machine learning?',
  documents: [
    'Machine learning is a subset of artificial intelligence.',
    'The weather is nice today.',
    'Deep learning uses neural networks.'
  ],
  top_n: 2
})

console.log(result.results)
```

## 测试

### 单元测试

已实现的测试文件：
- `internal/objects/rerank_test.go`: 测试数据结构的 JSON 序列化

运行测试：

```bash
go test ./internal/objects/... -v -run TestRerank
```

### 集成测试

要测试完整的 Rerank 功能流程，需要：

1. 启动 AxonHub 服务器
2. 配置至少一个支持 Rerank 的 Channel
3. 使用有效的 API Key 发送请求

## 注意事项

1. **Channel 支持**: 并非所有 Transformer 都实现了 `Transformer` 接口。只有实现了该接口的 Channel 才能使用 Rerank 功能。

2. **模型支持**: Rerank 功能依赖于上游 AI 提供商的模型支持。确保选择的模型支持 Rerank 操作。

3. **认证**: Rerank 端点需要 API Key 认证，与其他 API 端点一致。

4. **错误处理**: 如果 Channel 不支持 Rerank 或模型不可用，会返回适当的错误信息。

## 后续改进

1. **权限控制**: 当前实现了基本的 Channel 选择，未来可以添加更细粒度的权限控制。

2. **使用量记录**: 可以扩展以记录 Rerank 请求的使用量统计。

3. **更多提供商**: 目前主要支持 OpenAI 格式，可以添加其他提供商的 Rerank 实现（如 Cohere, Anthropic 等）。

4. **缓存**: 对于相同的查询和文档，可以考虑添加缓存机制提高性能。

## 文件清单

新增和修改的文件：

```
internal/objects/rerank.go                      # 数据结构定义
internal/objects/rerank_test.go                 # 数据结构测试
internal/llm/transformer/interfaces.go          # Transformer 接口扩展
internal/llm/transformer/openai/rerank.go       # OpenAI Rerank 实现
internal/server/biz/rerank.go                   # 业务逻辑层
internal/server/biz/fx_module.go                # 依赖注入配置 (修改)
internal/server/api/rerank.go                   # API 处理器
internal/server/api/fx_module.go                # 依赖注入配置 (修改)
internal/server/routes.go                       # 路由注册 (修改)
frontend/src/lib/api-client.ts                  # 前端 API 客户端 (修改)
```

## 架构图

```
Client Request
     |
     v
API Layer (internal/server/api/rerank.go)
     |
     v
Business Layer (internal/server/biz/rerank.go)
     |
     v
Transformer Layer (internal/llm/transformer/openai/rerank.go)
     |
     v
HTTP Request to Provider
     |
     v
Provider Response
     |
     v
Transform to Unified Format
     |
     v
Return to Client
```
