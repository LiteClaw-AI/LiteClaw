# 🚀 LiteClaw - 高性能 AI Agent CLI 工具

<div align="center">

**轻量级、易用、功能强大的 AI Agent CLI 工具**

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/Performance-Optimized-green.svg)](docs/OPTIMIZATION.md)

[English](#english) | [简体中文](#简体中文)

</div>

---

## 简体中文

### 📖 项目简介

LiteClaw 是一个使用 Go 语言实现的高性能 AI Agent CLI 工具，提供：

- ✅ **OpenAI 兼容 API 服务端** - 可作为 Cherry Studio 等工具的后端
- ✅ **多 Provider 支持** - 支持 20+ AI 服务商
- ✅ **RAG 系统** - 完整的文档检索增强生成能力
- ✅ **Agent 框架** - 工具调用、思维链、多步推理
- ✅ **向量数据库集成** - 支持 Qdrant, Pinecone, Milvus 等
- ✅ **极致性能优化** - 二进制 2.6 MB，内存 4.1 MB

### 🎯 核心特性

#### 1. 多 Provider 支持

支持 20+ AI 服务商：

**国际服务商**
- OpenAI (GPT-4, GPT-4 Turbo)
- Anthropic (Claude 3.5 Sonnet)
- Google (Gemini 1.5 Flash)
- Groq (Llama 3.3 70B - 极速推理)
- Mistral AI (Mistral Large)
- xAI (Grok 2)
- Cohere (Command R+)
- OpenRouter (模型聚合器)
- Together AI (开源模型)

**中国服务商**
- 阿里云通义千问
- 百度文心一言
- DeepSeek
- 月之暗面 Kimi

**本地部署**
- Ollama (本地模型)
- LocalAI (OpenAI 兼容)

#### 2. Agent 框架

完整的 Agent 能力：

- **工具调用** - 支持 Function Calling
- **思维链** - Chain of Thought 推理
- **多步推理** - 自动规划与执行
- **记忆管理** - 短期/长期记忆
- **工具生态** - Bash, 文件操作, Web 搜索等

#### 3. RAG 系统

生产级 RAG 能力：

- **文档加载** - PDF, Word, Markdown, Code
- **智能分块** - 递归、语义、句子分块
- **向量化** - 支持多种 Embedding 模型
- **向量存储** - 内存、Qdrant, Pinecone, Milvus, pgvector
- **检索优化** - 混合检索、重排序

#### 4. 向量数据库

支持多种向量数据库：

- **Memory** - 内存存储（开发测试）
- **Qdrant** - 高性能开源向量库
- **Pinecone** - 云托管向量服务
- **Milvus** - 云原生向量数据库
- **pgvector** - PostgreSQL 扩展

### 📊 性能指标

| 指标 | 数值 | 对比 Rust |
|------|------|----------|
| **二进制大小** | 2.6 MB | ✅ 小 7% |
| **内存占用** | 4.1 MB | ✅ 省 9% |
| **启动时间** | 3.5 ms | ✅ 快 42% |
| **编译速度** | 2 s | ✅ 快 96% |

### 🚀 快速开始

#### 安装

```bash
# 方式 1: 下载预编译二进制
# macOS/Linux
curl -fsSL https://get.liteclaw.dev | sh

# Windows
irm https://get.liteclaw.dev/windows | iex

# 方式 2: 从源码编译
git clone https://github.com/liteclaw/liteclaw
cd liteclaw
./scripts/quick-opt.sh
```

#### 基础使用

```bash
# Chat 模式
liteclaw chat --provider openai --model gpt-4

# RAG 模式
liteclaw rag index ./documents --collection my-docs
liteclaw rag query "什么是机器学习?" --collection my-docs

# API 服务器
liteclaw server --port 8080

# Agent 模式
liteclaw agent run --task "分析这份代码并给出优化建议"
```

#### 环境变量配置

```bash
# AI Provider 配置
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GROQ_API_KEY="gsk_..."

# 中国服务商
export DEEPSEEK_API_KEY="..."
export ALIYUN_API_KEY="..."

# 本地模型
export OLLAMA_BASE_URL="http://localhost:11434"
```

### 📖 使用示例

#### 1. 作为 API 服务端

```bash
# 启动服务器
liteclaw server --port 8080 --provider openai

# 在 Cherry Studio 中配置
# API Base: http://localhost:8080/v1
# API Key: 任意值
# Model: gpt-4
```

#### 2. 使用 RAG

```bash
# 索引文档
liteclaw rag index ./docs --collection knowledge-base

# 查询
liteclaw rag query "如何优化性能?" --collection knowledge-base

# 增量更新
liteclaw rag update ./new-docs --collection knowledge-base
```

#### 3. 使用 Agent

```bash
# 运行 Agent 任务
liteclaw agent run --task "分析项目代码并生成文档"

# 使用工具
liteclaw agent run --task "读取 README.md 并总结" --tools file_read

# 多步任务
liteclaw agent run --task "搜索 Go 性能优化最佳实践，并生成报告"
```

#### 4. 自定义 Provider

```go
package main

import (
    "context"
    "fmt"
    "liteclaw/pkg/provider"
)

func main() {
    // 创建自定义 Provider
    customProvider := provider.NewOpenAI("your-api-key")
    
    // Chat
    resp, err := customProvider.Chat(context.Background(), &provider.ChatRequest{
        Messages: []provider.Message{
            {Role: provider.RoleUser, Content: "Hello!"},
        },
        Model: "gpt-4",
    })
    
    fmt.Println(resp.Choices[0].Message.Content)
}
```

### 🏗️ 项目结构

```
.
├── cmd/                    # CLI 命令
│   ├── chat.go            # Chat 命令
│   ├── rag.go             # RAG 命令
│   ├── server.go          # Server 命令
│   └── root.go            # 根命令
├── pkg/
│   ├── provider/          # AI Provider 系统
│   │   ├── openai.go      # OpenAI
│   │   ├── anthropic.go   # Anthropic
│   │   ├── groq.go        # Groq
│   │   ├── mistral.go     # Mistral
│   │   ├── gemini.go      # Gemini
│   │   ├── china.go       # 中国厂商
│   │   └── init.go        # 自动注册
│   ├── rag/               # RAG 系统
│   │   ├── loader.go      # 文档加载
│   │   ├── splitter.go    # 文本分块
│   │   ├── vector.go      # 向量存储
│   │   └── pipeline.go    # RAG 流程
│   ├── agent/             # Agent 框架
│   │   ├── agent.go       # Agent 定义
│   │   ├── executor.go    # 执行引擎
│   │   ├── memory.go      # 记忆管理
│   │   ├── tools.go       # 工具集
│   │   └── workflow.go    # 工作流
│   ├── vectordb/          # 向量数据库
│   │   ├── store.go       # 接口定义
│   │   ├── memory.go      # 内存存储
│   │   ├── qdrant.go      # Qdrant
│   │   ├── pinecone.go    # Pinecone
│   │   ├── milvus.go      # Milvus
│   │   └── pgvector.go    # pgvector
│   └── api/               # API 服务
│       └── server.go      # Gin 服务器
├── scripts/               # 构建脚本
│   ├── quick-opt.sh       # 快速优化
│   └── build-optimized.sh # 完整构建
├── docs/                  # 文档
│   ├── OPTIMIZATION.md    # 优化指南
│   └── RUST_VS_GO.md      # 性能对比
├── examples/              # 示例代码
├── main.go                # 主入口
├── optimize.go            # 优化配置
└── README.md              # 本文档
```

### 🔧 高级功能

#### Agent 工作流

```go
// 创建 Agent
agent := agent.NewAgent(
    "code-analyzer",
    "Code Analyzer",
    "Analyzes code and provides suggestions",
    agent.WithProvider("openai"),
    agent.WithModel("gpt-4"),
    agent.WithTools(
        tools.NewFileReadTool(),
        tools.NewBashTool(),
    ),
    agent.WithConfig(agent.AgentConfig{
        MaxIterations:  10,
        Temperature:    0.7,
        EnableMemory:   true,
        EnableThinking: true,
    }),
)

// 执行任务
executor := agent.NewExecutor(providerRegistry)
result, err := executor.Execute(ctx, agent, "分析 main.go 并给出优化建议")
```

#### 向量数据库集成

```go
// 创建 Qdrant 客户端
store, err := vectordb.NewVectorStore("qdrant", vectordb.Config{
    URL:            "http://localhost:6333",
    Collection:     "documents",
    Dimension:      1536,
    DistanceMetric: "cosine",
})

// 插入向量
err = store.Insert(ctx, "documents", []vectordb.Vector{
    {
        ID:     "doc1",
        Values: embedding,
        Metadata: map[string]interface{}{
            "title":  "Document 1",
            "source": "file.pdf",
        },
    },
})

// 搜索
results, err := store.Search(ctx, "documents", queryVector, 10, nil)
```

### 📚 文档

- [优化指南](docs/OPTIMIZATION.md) - 性能优化详细说明
- [优化报告](docs/OPTIMIZATION_REPORT.md) - 优化过程记录
- [Rust vs Go](docs/RUST_VS_GO.md) - 详细性能对比
- [最终报告](FINAL_OPTIMIZATION_REPORT.md) - 最终优化结果

### 🛠️ 开发

```bash
# 开发模式
go run main.go chat --provider openai

# 测试
go test ./...

# 构建
go build -o liteclaw

# 优化构建
./scripts/quick-opt.sh

# Docker
docker build -f Dockerfile.opt -t liteclaw:opt .
```

### 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

### 📝 License

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## English

### 📖 Introduction

LiteClaw is a high-performance AI Agent CLI tool implemented in Go, featuring:

- ✅ **OpenAI-compatible API Server** - Backend for Cherry Studio and similar tools
- ✅ **Multi-Provider Support** - 20+ AI service providers
- ✅ **RAG System** - Complete Retrieval-Augmented Generation
- ✅ **Agent Framework** - Tool calling, chain of thought, multi-step reasoning
- ✅ **Vector Database Integration** - Qdrant, Pinecone, Milvus, etc.
- ✅ **Extreme Performance** - 2.6 MB binary, 4.1 MB memory

### 🚀 Quick Start

```bash
# Install
curl -fsSL https://get.liteclaw.dev | sh

# Chat mode
liteclaw chat --provider openai --model gpt-4

# RAG mode
liteclaw rag index ./documents --collection my-docs
liteclaw rag query "What is machine learning?" --collection my-docs

# API server
liteclaw server --port 8080
```

### 📊 Performance

| Metric | Value | vs Rust |
|--------|-------|---------|
| **Binary Size** | 2.6 MB | ✅ 7% smaller |
| **Memory Usage** | 4.1 MB | ✅ 9% less |
| **Startup Time** | 3.5 ms | ✅ 42% faster |
| **Build Speed** | 2 s | ✅ 96% faster |

### 📚 Documentation

- [Optimization Guide](docs/OPTIMIZATION.md)
- [Optimization Report](docs/OPTIMIZATION_REPORT.md)
- [Rust vs Go](docs/RUST_VS_GO.md)
- [Final Report](FINAL_OPTIMIZATION_REPORT.md)

### 📝 License

MIT License - See [LICENSE](LICENSE)

---

<div align="center">

**Made with ❤️ by LiteClaw Team**

[Website](https://liteclaw.dev) | [Documentation](https://docs.liteclaw.dev) | [GitHub](https://github.com/liteclaw/liteclaw)

</div>
