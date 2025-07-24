# Generation Blog

[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://golang.org/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-brightgreen)](https://vuejs.org/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-orange)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/Redis-6.0+-red)](https://redis.io/)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-7.x-yellow)](https://www.elastic.co/)

> 一个由 Go 语言后端 + Vue 前端构成的现代化博客平台，集成 AI 智能分析、搜索推荐、多数据源同步等功能。

## 📖 项目简介

Generation Blog 是一个功能完整的全栈博客项目，历时两个月开发完成。从 API 设计、数据库建模，到用户身份认证、文章发布、评论系统，再到搜索推荐、标签管理等一整套功能的实现。

这个博客承载了对"技术表达"的思考：构建一个既快速、稳定，又结构清晰的系统，让写作回归简单纯粹。

### 💫 核心特性

- **🏗️ 前后端分离架构**：Go Gin 后端 + Vue 前端，配套用户端和管理系统
- **🔍 智能搜索**：基于 Elasticsearch，支持全文搜索和段落级检索
- **🤖 AI 深度集成**：ChatGPT 驱动的文章分析、标题生成、智能对话
- **📊 高性能数据架构**：MySQL 主从复制、Redis 缓存、垂直分表优化
- **🎯 论文自动生成**：ArXiv 爬虫 + AI 分析，自动生成高质量技术文章
- **☁️ 对象存储**：前后端分离的文件上传，支持直传和权限控制
- **🔐 安全认证**：JWT + Redis 黑名单机制
- **📱 实时同步**：binlog 监听实现数据实时同步

## 🚀 技术亮点

### 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue 前端       │◄──►│   Gin API 网关   │◄──►│   核心业务层     │
│                 │    │                 │    │                 │
│ • 用户界面       │    │ • 路由分发       │    │ • 文章管理       │
│ • 管理后台       │    │ • 中间件链       │    │ • 用户系统       │
│ • AI 对话        │    │ • 参数验证       │    │ • 评论系统       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                ▲
                                │
         ┌──────────────────────┼──────────────────────┐
         ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   数据存储层     │    │   搜索引擎       │    │   缓存层         │
│                 │    │                 │    │                 │
│ • MySQL 主从    │    │ • Elasticsearch │    │ • Redis 缓存    │
│ • 读写分离      │    │ • 实时同步      │    │ • 会话管理      │
│ • 垂直分表      │    │ • 全文检索      │    │ • AI 结果缓存   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 数据库架构

- **MySQL 主从复制**：读写分离，提升并发能力
- **Redis 缓存策略**：热点数据缓存，会话状态管理
- **Elasticsearch 同步**：binlog 监听，实时数据同步
- **垂直分表优化**：按业务模块分表，提升查询性能

### AI 智能功能

- **文章智能分析**：自动生成标题、摘要、关键词标签
- **论文爬虫系统**：ArXiv 多领域论文自动爬取和分析
- **对话式搜索**：结合 Elasticsearch 推荐引擎，实现"对话即搜索"
- **批量评分算法**：双重随机批次评分，提升 AI 分析效率

## 🛠️ 技术栈

### 后端技术

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.24 | 核心开发语言 |
| Gin | 1.10+ | Web 框架 |
| GORM | 1.30+ | ORM 框架 |
| MySQL | 8.0 | 主数据库 |
| Redis | 6.0+ | 缓存和会话 |
| Elasticsearch | 7.x | 搜索引擎 |
| JWT | - | 身份认证 |
| Docker | - | 容器化部署 |

### 第三方服务

- **AI 服务**：ChatGPT 3.5 (计划迁移至 DeepSeek v3)
- **对象存储**：七牛云存储
- **邮件服务**：SMTP 邮件发送
- **IP 定位**：ip2region 库

## 📦 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+
- Docker & Docker Compose (可选)

### 1. 克隆项目

```bash
git clone <repository-url>
cd blogX_server
```

### 2. 配置文件

复制配置模板并修改：

```bash
cp settings.yaml.example settings.yaml
```

编辑 `settings.yaml` 文件，配置数据库、Redis、AI 等服务：

```yaml
system:
  ip: "0.0.0.0"
  port: 8080
  env: "prod"
  gin_mode: "release"

db:
  - name: "master"
    user: "root"
    password: "your_password"
    host: "localhost"
    port: 3306
    dbname: "blogx"
    debug: false
    source: "mysql"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

ai:
  enable: true
  secretKey: "your_openai_api_key"
  chatModel: "gpt-3.5-turbo"
```

### 3. 数据库初始化

```bash
# 安装依赖
go mod tidy

# 数据库迁移
go run main.go -db

# 创建管理员用户
go run main.go -t user -s create
```

### 4. 启动服务

```bash
# 开发环境
go run main.go

# 生产环境
go build -o blogx_server
./blogx_server
```

### 5. Docker 部署 (推荐)

```bash
# 构建镜像
docker build -t blogx_server:v1 .

# 使用 docker-compose 启动完整环境
cd init/deploy
docker-compose up -d
```

服务启动后访问：http://localhost:8080

## 🔧 配置说明

### 核心配置项

#### 系统配置
```yaml
system:
  ip: "0.0.0.0"        # 监听地址
  port: 8080           # 端口号
  env: "prod"          # 环境：dev/prod/test
  gin_mode: "release"  # Gin 模式
```

#### 数据库配置
```yaml
db:
  - name: "master"     # 主库
    user: "root"
    password: "password"
    host: "localhost"
    port: 3306
    dbname: "blogx"
  - name: "slave"      # 从库（可选）
    user: "root"
    password: "password"
    host: "slave-host"
    port: 3306
    dbname: "blogx"
```

#### AI 配置
```yaml
ai:
  enable: true
  secretKey: "your_openai_api_key"
  nickname: "AI助手"
  chatModel: "gpt-3.5-turbo"
  thinkModel: "gpt-4o-mini"
```

#### 搜索配置
```yaml
es:
  addr: "localhost:9200"
  isHttps: false
  username: ""
  password: ""

river:
  enable: true
  serverID: 1001
  flavor: "mysql"
  dataDir: "./var"
```

### 站点功能配置

```yaml
site:
  siteInfo:
    title: "Generation Blog"
    mode: 1  # 1-社区模式 2-博客模式
  autoGen:
    userID: 1
    categories: ["cs.AI", "astro-ph", "quant-ph"]
    limit: 200
    top: 20
```

## 📚 API 文档

### 用户相关接口

#### 注册登录
```http
POST /api/user/register_email    # 邮箱注册
POST /api/user/pwd_login         # 密码登录
POST /api/user/logout            # 用户登出
```

#### 用户管理
```http
GET  /api/user/detail            # 用户详情
PUT  /api/user/info_update       # 更新用户信息
POST /api/user/change_password   # 修改密码
```

### 文章相关接口

#### 文章管理
```http
GET    /api/article/list         # 文章列表
POST   /api/article/create       # 创建文章
GET    /api/article/detail/:id   # 文章详情
PUT    /api/article/update/:id   # 更新文章
DELETE /api/article/remove       # 删除文章
```

#### 文章操作
```http
POST /api/article/like/:id       # 点赞文章
POST /api/article/collect/:id    # 收藏文章
POST /api/article/read/:id       # 标记已读
```

### 评论相关接口

```http
GET  /api/comment/list/:id       # 评论列表
POST /api/comment/create         # 创建评论
POST /api/comment/like/:id       # 点赞评论
DELETE /api/comment/remove/:id   # 删除评论
```

### 搜索相关接口

```http
GET /api/search/article          # 文章搜索
GET /api/search/text             # 全文搜索
GET /api/search/tag_agg          # 标签聚合
```

### AI 相关接口

```http
POST /api/ai/chat                # AI 对话
POST /api/ai/article_analysis    # 文章分析
```

## 🧠 AI 功能详解

### 1. 论文自动生成系统

Generation Blog 集成了强大的论文分析和文章自动生成系统：

#### ArXiv 爬虫服务
- 支持 7 个主要学科领域（AI、天体物理、量子物理等）
- 实时爬取最新论文，每领域 150-230 篇
- 自动识别论文类别，提供中文分类名称

#### 智能批量评分算法
- **双重随机批次评分**：每篇论文分配到 2 个不同批次
- **相对评分机制**：批次内论文相对比较，提升区分度
- **冲突检测**：自动识别评分差异，触发第三轮评分
- **分项评分**：创新性(40分) + 技术深度(30分) + 实用性(30分)

#### 两阶段分析流程
```
Stage 1: 批次评分阶段
├── 论文随机分配到批次
├── 并行执行批次评分
├── 冲突检测和第三轮评分
└── 分数合并

Stage 2: 详细分析阶段
├── 选择 Top-N 高分论文
├── 并行生成中文摘要
├── 专业评价和关键词提取
└── 生成完整分析报告
```

### 2. AI 对话助手

- **上下文感知**：记忆对话历史，提供连贯回复
- **文章推荐**：结合 Elasticsearch，精准推送相关文章
- **多轮对话**：支持复杂问题的分步解答

### 3. 文章智能分析

- **自动标题生成**：基于内容生成吸引人的标题
- **摘要提取**：智能提取文章核心要点
- **标签推荐**：自动生成相关标签和关键词
- **质量评估**：评估文章创新性和技术价值

## 🔍 搜索系统

### Elasticsearch 集成

- **实时同步**：监听 MySQL binlog，实时同步数据变更
- **全文检索**：支持中文分词，精确到段落级别
- **聚合分析**：按标签、类别、时间等维度聚合统计
- **高性能**：复杂查询毫秒级响应

### 搜索功能

- **文章搜索**：标题、内容、标签全文搜索
- **用户搜索**：按用户名、昵称搜索
- **标签搜索**：相关标签智能推荐
- **高级搜索**：时间范围、分类筛选等

## 🗄️ 数据模型

### 核心数据表

#### 用户系统
- `user_model` - 用户基本信息
- `user_config_model` - 用户配置
- `user_login_model` - 登录历史
- `user_focus_model` - 关注关系

#### 文章系统
- `article_model` - 文章内容
- `category_model` - 文章分类
- `article_likes_model` - 文章点赞
- `user_pinned_article_model` - 置顶文章

#### 评论系统
- `comment_model` - 评论内容
- `comment_likes_model` - 评论点赞

#### 收藏系统
- `collection_folder_model` - 收藏夹
- `article_collection_model` - 文章收藏

#### 系统管理
- `log_model` - 操作日志
- `global_notification_model` - 全局通知
- `data_model` - 统计数据

## 🔐 安全机制

### 身份认证
- **JWT Token**：无状态认证，支持分布式部署
- **Redis 黑名单**：token 吊销机制，增强安全性
- **多级权限**：超级管理员、普通用户、访客三级权限

### 数据安全
- **SQL 注入防护**：使用 GORM 预处理语句
- **XSS 防护**：输入输出过滤和转义
- **CSRF 防护**：验证请求来源
- **参数验证**：严格的输入参数校验

### 系统安全
- **密码哈希**：bcrypt 加密存储
- **敏感信息加密**：配置文件敏感数据加密
- **访问日志**：完整的操作审计日志

## 📊 运维监控

### 日志系统
- **分级日志**：Error、Warn、Info、Debug 四级日志
- **结构化日志**：JSON 格式，便于分析
- **日志轮转**：按时间和大小自动轮转
- **远程日志**：支持发送到外部日志系统

### 性能监控
- **系统状态**：CPU、内存、磁盘使用率
- **数据库监控**：连接数、慢查询统计
- **缓存监控**：Redis 连接状态和命中率
- **业务指标**：用户活跃度、文章发布量等

### 定时任务
- **数据同步**：定时同步缓存数据到数据库
- **文章生成**：定时执行论文分析和文章生成
- **数据清理**：清理过期缓存和临时文件
- **统计更新**：更新网站统计数据

## 🚀 部署指南

### Docker 部署

#### 1. 准备环境
```bash
# 克隆项目
git clone <repository-url>
cd blogX_server/init/deploy

# 修改配置文件
cp blogx_server/settings.yaml.example blogx_server/settings.yaml
```

#### 2. 构建镜像
```bash
# 构建后端镜像
docker build -t blogx_server:v1 ../../

# 准备数据库初始化文件
# 将现有数据库导出为 master/blogx.sql
```

#### 3. 启动服务
```bash
# 启动完整环境
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f blogx_server
```

### 传统部署

#### 1. 环境安装
```bash
# 安装 Go 1.24+
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz

# 安装 MySQL 8.0
# 安装 Redis 6.0+
# 安装 Elasticsearch 7.x (可选)
```

#### 2. 编译部署
```bash
# 克隆代码
git clone <repository-url>
cd blogX_server

# 编译
go mod tidy
go build -o blogx_server

# 配置
cp settings.yaml.example settings.yaml
# 编辑 settings.yaml

# 初始化数据库
./blogx_server -db

# 启动服务
./blogx_server
```

#### 3. 反向代理 (Nginx)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location / {
        root /path/to/vue/dist;
        try_files $uri $uri/ /index.html;
    }
}
```

## 🔧 开发指南

### 项目结构
```
blogX_server/
├── api/                    # API 接口层
│   ├── article_api/       # 文章相关接口
│   ├── user_api/          # 用户相关接口
│   ├── comment_api/       # 评论相关接口
│   └── ...
├── common/                # 公共工具
├── conf/                  # 配置管理
├── core/                  # 核心初始化
├── flags/                 # 命令行参数
├── global/                # 全局变量
├── middleware/            # 中间件
├── models/                # 数据模型
├── router/                # 路由配置
├── service/               # 业务逻辑层
│   ├── ai_service/        # AI 服务
│   ├── article_auto_generate/ # 文章自动生成
│   ├── redis_service/     # Redis 服务
│   └── ...
└── utils/                 # 工具函数
```

### 开发环境配置

1. **IDE 配置**：推荐使用 GoLand 或 VS Code
2. **代码规范**：使用 gofmt 和 golint
3. **Git 规范**：使用 conventional commits
4. **测试**：编写单元测试和集成测试

### 扩展开发

#### 添加新的 API 接口
1. 在 `api/` 目录下创建新的接口文件
2. 在 `router/` 中注册路由
3. 在 `service/` 中实现业务逻辑
4. 更新文档和测试

#### 添加新的数据模型
1. 在 `models/` 中定义新的结构体
2. 在 `flags/flag_db.go` 中添加迁移
3. 实现相关的业务逻辑

## ❓ 常见问题

### Q: 如何配置 AI 功能？
A: 在 `settings.yaml` 中配置 `ai.secretKey`，目前支持 OpenAI API。

### Q: Elasticsearch 是否必需？
A: 不是必需的。如果不配置 ES，搜索功能会降级为数据库查询。

### Q: 如何添加新的论文分类？
A: 在 `service/article_auto_generate/crawler_service/` 中添加新的分类常量和爬取逻辑。

### Q: 如何自定义 AI 提示词？
A: 修改 `service/ai_service/` 目录下的 `.prompt` 文件。

### Q: 如何配置对象存储？
A: 在 `settings.yaml` 中配置 `cloud.qny` 相关参数，支持七牛云存储。

## 🛠️ 故障排除

### 常见错误

#### 数据库连接失败
```
检查 settings.yaml 中的数据库配置
确认 MySQL 服务正在运行
检查网络连接和防火墙设置
```

#### Redis 连接失败
```
检查 Redis 服务状态
确认配置中的地址和端口正确
检查 Redis 密码配置
```

#### AI 功能异常
```
检查 API Key 配置是否正确
确认网络可以访问 OpenAI API
查看日志中的具体错误信息
```

### 性能优化

#### 数据库优化
- 为查询频繁的字段添加索引
- 使用读写分离减轻主库压力
- 定期清理过期数据

#### 缓存优化
- 合理设置 Redis 过期时间
- 使用缓存预热提升响应速度
- 监控缓存命中率

## 🤝 贡献指南

欢迎贡献代码、报告问题或提出建议！

### 如何贡献

1. **Fork 项目**
2. **创建特性分支** (`git checkout -b feature/amazing-feature`)
3. **提交更改** (`git commit -m 'Add some amazing feature'`)
4. **推送分支** (`git push origin feature/amazing-feature`)
5. **创建 Pull Request**

### 代码规范

- 遵循 Go 代码规范
- 添加必要的注释
- 编写单元测试
- 更新相关文档

### 报告问题

使用 GitHub Issues 报告 Bug 或提出功能请求。请提供：
- 详细的问题描述
- 复现步骤
- 环境信息
- 错误日志

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢所有使用和贡献本项目的开发者！

特别感谢以下开源项目：
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://github.com/go-gorm/gorm) - Go ORM 库
- [Redis](https://redis.io/) - 内存数据库
- [Elasticsearch](https://www.elastic.co/) - 搜索引擎
- [Vue.js](https://vuejs.org/) - 前端框架

## 🔗 相关链接

- **项目主页**: [Generation Blog](https://your-blog-url.com)
- **API 文档**: [API Documentation](https://your-api-docs.com)
- **前端项目**: [Vue Frontend Repository](https://github.com/your-frontend-repo)
- **作者博客**: [LIR's Blog](https://your-personal-blog.com)

---

*Generation Blog - 让技术表达回归简单纯粹* 