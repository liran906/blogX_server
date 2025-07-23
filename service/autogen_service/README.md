# AI论文分析服务 - 接口调用指南

## 📖 概述

本文档介绍如何使用AI论文分析服务来获取、分析ArXiv论文并生成文章内容。

## 🏗️ 架构优势

### **新设计理念**
- **爬虫数据**：成本低，实时获取，不缓存
- **AI分析结果**：成本高，智能缓存7天，避免重复分析

### **核心优势**
✅ **成本节约** - AI分析结果缓存，避免重复API调用  
✅ **数据新鲜** - 论文数据实时爬取，始终最新  
✅ **智能缓存** - 自动管理缓存生命周期，7天过期  
✅ **多类别支持** - 支持AI、量子物理、天体物理等多个领域  
✅ **高性能** - 缓存命中时响应极快

## 🚀 快速开始

### 1. 初始化服务

```go
package main

import (
    "fmt"
    "log"
    "time"
    "strings"
    "blogX_server/service/autogen_service"
    "blogX_server/service/crawler_service"
    "blogX_server/core"
    "blogX_server/global"
    "blogX_server/flags"
)

func main() {
    // 初始化配置（必须）
    flags.Parse()
    global.Config = core.ReadConf()
    core.InitLogrus()
    global.Redis = core.InitRedis()
    
    // 创建AI分析服务
    service := autogen_service.NewAutogenService()
    
    // 调用分析接口
    generateArticles(service)
}
```

### 2. 主要接口调用

```go
func generateArticles(service *autogen_service.AutogenService) {
    // 实时爬取AI类别论文，限制50篇，筛选出评分最高的5篇用于写文章
    topPapers, err := service.AnalyzePapersForWriting(crawler_service.CategoryAI, 50, 5)
    if err != nil {
        log.Printf("获取论文失败: %v", err)
        return
    }
    
    if len(topPapers) == 0 {
        log.Printf("没有找到合适的论文")
        return
    }
    
    // 遍历高分论文，生成文章
    for i, paper := range topPapers {
        fmt.Printf("=== 准备写作第 %d 篇文章 ===\n", i+1)
        
        // 你可以基于这些数据生成文章
        articleData := ArticleData{
            Title:         generateTitle(paper),
            Content:       generateContent(paper), 
            Tags:          paper.Tags,
            Score:         paper.Score,
            SourcePaper:   paper.Title,
            SourceAuthors: paper.Authors,
            SourceURL:     paper.HtmlURL,  // 使用HTML链接
            PdfURL:        paper.PdfURL,   // PDF下载链接
            PublishTime:   time.Now(),
        }
        
        // 调用你的文章发布逻辑
        publishArticle(articleData)
    }
}
```

## 📋 数据结构

### PaperAnalysisResult 论文分析结果

```go
type PaperAnalysisResult struct {
    ArxivID          string   `json:"arxivId"`          // 原始ArXiv ID
    Title            string   `json:"title"`            // 论文标题
    Authors          string   `json:"authors"`          // 作者列表  
    PublishedDate    string   `json:"publishedDate"`    // 发表时间
    Abstract         string   `json:"abstract"`         // AI生成的中文摘要
    Score            int      `json:"score"`            // 科研价值评分(0-100)
    Justification    string   `json:"just"`             // 评分理由
    Tags             []string `json:"tags"`             // 主题标签
    AnalyzedAt       string   `json:"analyzedAt"`       // 分析时间
    OriginalAbstract string   `json:"originalAbstract"` // 原始英文摘要
    PdfURL           string   `json:"pdfUrl"`           // PDF链接
    HtmlURL          string   `json:"htmlUrl"`          // HTML链接
}
```

### ArticleData 文章数据结构（示例）

```go
type ArticleData struct {
    Title         string    // 文章标题
    Content       string    // 文章内容
    Tags          []string  // 标签
    Score         int       // 原论文评分
    SourcePaper   string    // 原论文标题
    SourceAuthors string    // 原论文作者
    SourceURL     string    // 原论文HTML链接
    PdfURL        string    // 原论文PDF链接
    PublishTime   time.Time // 发布时间
}
```

## 🎯 核心接口

### AnalyzePapersForWriting

**功能**: 实时爬取指定类别论文，AI分析评分，返回排序后的高分论文（AI分析结果自动缓存7天）

**签名**:
```go
func (s *AutogenService) AnalyzePapersForWriting(category crawler_service.ArxivCategory, limit int, topN int) ([]*PaperAnalysisResult, error)
```

**参数**:
- `category`: 论文类别（如 CategoryAI、CategoryQuantumPhysics 等）
- `limit`: 爬取论文数量上限
- `topN`: 返回评分最高的N篇论文

**返回**: 按评分降序排列的论文分析结果，包含PDF和HTML链接

**示例**:
```go
// 爬取AI类别100篇论文，选出评分最高的5篇
topPapers, err := service.AnalyzePapersForWriting(crawler_service.CategoryAI, 100, 5)

// 爬取量子物理类别50篇论文，选出评分最高的3篇
quantumPapers, err := service.AnalyzePapersForWriting(crawler_service.CategoryQuantumPhysics, 50, 3)

// 访问论文链接
for _, paper := range topPapers {
    fmt.Printf("论文: %s\n", paper.Title)
    fmt.Printf("在线阅读: %s\n", paper.HtmlURL)
    fmt.Printf("PDF下载: %s\n", paper.PdfURL)
}
```

### 其他接口

```go
// 批量分析（不排序）
results, err := service.AnalyzePapers(papers)

// 获取排序后的论文（自定义数量）
topPapers := autogen_service.GetTopScoredPapers(results, 10)

// 生成分析报告
report := service.GenerateAnalysisReport(results)
```

### FormatAnalysisReport

**功能**: 将AI分析结果格式化为美观的Markdown报告

**签名**:
```go
func FormatAnalysisReport(results []*PaperAnalysisResult) string
```

**参数**:
- `results`: 论文分析结果数组

**返回**: 格式化的Markdown报告字符串

**输出格式特点**:
- 📊 **报告头部**: 包含生成时间、分析数量、评分依据
- ⭐ **评分概览**: 显示最高分、平均分、最低分统计
- 📄 **详细分析**: 每篇论文的完整信息
- 🎯 **智能评分**: 根据分数显示不同表情符号（🔥⭐👍👌📝）
- 🏷️ **标签展示**: 关键词用代码格式美化
- 🔗 **链接整合**: ArXiv和PDF链接并排显示

**示例**:
```go
// 生成Markdown格式的分析报告
report := autogen_service.FormatAnalysisReport(topPapers)

// 保存到文件或发布到博客
ioutil.WriteFile("analysis_report.md", []byte(report), 0644)

// 或者直接在网页中展示
fmt.Print(report)
```

**输出样例**:
```markdown
## 📊 AI论文分析报告

📅 **生成时间**: 2025-07-23 22:34:01  
📚 **分析数量**: 3 篇论文  
🎯 **评分依据**: 创新性、技术难度、应用价值

### ⭐ 评分概览

- 🏆 **最高分**: `95`  
- 📊 **平均分**: `79.3`  
- 📉 **最低分**: `65`  

---

## 📄 详细分析

### Advanced Machine Learning Techniques for Natural Language Processing

👥 **作者**: John Smith, Jane Doe, Bob John...  
🕒 **分析时间**: 2025-07-23 22:34:01  
🔗 **论文源**: [📄 ArXiv](https://arxiv.org/abs/2507.15865) | [📥 PDF](https://arxiv.org/pdf/2507.15865.pdf)  
🎯 **科研评分**: `95/100` 🔥  
🔍 **本站分析**: 该论文在自然语言处理领域提出了创新性的深度学习方法...  
🏷️ **关键标签**: `机器学习` `自然语言处理` `深度学习`  
📝 **AI摘要**: 本研究提出了一种新颖的机器学习方法...

---
```

## 💡 实用工具函数

### 生成文章标题

```go
func generateTitle(paper *autogen_service.PaperAnalysisResult) string {
    if paper.Score >= 70 {
        return fmt.Sprintf("突破性研究：%s", paper.Title)
    } else if paper.Score >= 60 {
        return fmt.Sprintf("最新进展：%s", paper.Title) 
    } else {
        return fmt.Sprintf("研究解读：%s", paper.Title)
    }
}
```

### 生成文章内容

```go
func generateContent(paper *autogen_service.PaperAnalysisResult) string {
    content := fmt.Sprintf(`
## 论文概述
**标题**: %s
**作者**: %s  
**发表时间**: %s
**研究评分**: %d/100

## 核心亮点
%s

## 技术标签
%s

## 论文摘要
%s

## 我的点评
基于AI分析，这篇论文的%s，值得关注。

## 相关链接
- [📖 在线阅读](%s)
- [📄 PDF下载](%s)
`, 
        paper.Title,
        paper.Authors, 
        paper.PublishedDate,
        paper.Score,
        paper.Justification,
        strings.Join(paper.Tags, " | "),
        paper.Abstract,
        getScoreComment(paper.Score),
        paper.HtmlURL,
        paper.PdfURL,
    )
    
    return content
}
```

### 评分评语

```go
func getScoreComment(score int) string {
    if score >= 80 {
        return "创新性和实用价值都很高"
    } else if score >= 70 {
        return "具有较好的研究价值"
    } else if score >= 60 {
        return "有一定的参考意义" 
    } else {
        return "作为基础研究有其价值"
    }
}
```

### 发布文章

```go
func publishArticle(article ArticleData) {
    fmt.Printf("📝 发布文章: %s\n", article.Title)
    fmt.Printf("🏷️ 标签: %v\n", article.Tags)
    fmt.Printf("⭐ 评分: %d\n", article.Score)
    fmt.Printf("🔗 源链接: %s\n", article.SourceURL)
    fmt.Printf("📄 PDF: %s\n", article.PdfURL)
    
    // TODO: 调用你的文章发布API
    // 比如保存到数据库、推送到前端等
}
```

## ⚠️ 错误处理

```go
topPapers, err := service.AnalyzePapersForWriting(50, 5)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "论文列表为空"):
        log.Printf("需要先爬取论文数据")
        return
    case strings.Contains(err.Error(), "AI分析失败"):
        log.Printf("AI服务异常，请检查API配置")
        return
    default:
        log.Printf("未知错误: %v", err)
        return
    }
}
```

## 🔗 链接格式说明

### ArXiv链接规则

对于ArXiv ID为 `2024.12345v1` 的论文:

- **HTML链接**: `https://arxiv.org/abs/2024.12345v1`
- **PDF链接**: `https://arxiv.org/pdf/2024.12345v1.pdf`

### 使用示例

```go
// 获取论文链接
paper := topPapers[0]
fmt.Printf("论文ID: %s\n", paper.ArxivID)
fmt.Printf("在线查看: %s\n", paper.HtmlURL)
fmt.Printf("下载PDF: %s\n", paper.PdfURL)

// 在文章中嵌入链接
articleContent := fmt.Sprintf(`
查看原文：[%s](%s)
下载PDF：[点击下载](%s)
`, paper.Title, paper.HtmlURL, paper.PdfURL)
```

## 🔄 推荐的工作流程

### 每日定时任务

```go
func dailyArticleGeneration() {
    // 1. 爬取最新论文
    crawler := crawler_service.NewArxivCrawler()
    err := crawler.CrawlPaperAbstract(crawler_service.CS_AI, 100)
    if err != nil {
        log.Printf("爬取失败: %v", err)
        return
    }
    
    // 2. AI分析并生成文章
    service := autogen_service.NewAutogenService() 
    topPapers, err := service.AnalyzePapersForWriting(100, 3)
    if err != nil {
        log.Printf("分析失败: %v", err)
        return
    }
    
    // 3. 发布文章（包含链接）
    for _, paper := range topPapers {
        article := generateArticleFromPaper(paper)
        publishArticle(article)
    }
}
```

### 手动触发流程

```go
func manualArticleGeneration(category crawler_service.ArxivCategory, limit, topN int) {
    // 1. 指定类别爬取
    crawler := crawler_service.NewArxivCrawler()
    crawler.CrawlPaperAbstract(category, limit)
    
    // 2. 分析筛选
    service := autogen_service.NewAutogenService()
    topPapers, _ := service.AnalyzePapersForWriting(limit, topN)
    
    // 3. 预览结果（包含链接）
    for i, paper := range topPapers {
        fmt.Printf("推荐 %d: %s (评分: %d)\n", i+1, paper.Title, paper.Score)
        fmt.Printf("写作角度: %v\n", paper.Tags)
        fmt.Printf("核心价值: %s\n", paper.Justification)
        fmt.Printf("HTML链接: %s\n", paper.HtmlURL)
        fmt.Printf("PDF链接: %s\n", paper.PdfURL)
        fmt.Println("---")
    }
}
```

## 🎯 实际使用示例

### 完整示例

```go
package main

import (
    "fmt"
    "log"
    "blogX_server/service/autogen_service"
    "blogX_server/service/crawler_service"
    "blogX_server/core"
    "blogX_server/global"
    "blogX_server/flags"
)

func main() {
    // 初始化
    flags.Parse()
    global.Config = core.ReadConf()
    core.InitLogrus()
    global.Redis = core.InitRedis()
    
    // 创建服务
    autogenService := autogen_service.NewAutogenService()
    
    // 方案1: 直接从Redis分析（推荐）
    directAnalysis(autogenService)
    
    // 方案2: 先爬取再分析
    // crawlAndAnalysis(autogenService)
}

func directAnalysis(service *autogen_service.AutogenService) {
    fmt.Println("=== 直接从Redis分析论文 ===")
    
    topPapers, err := service.AnalyzePapersForWriting(50, 3)
    if err != nil {
        log.Printf("分析失败: %v", err)
        return
    }
    
    for i, paper := range topPapers {
        fmt.Printf("\n【推荐 %d】%s (评分: %d)\n", i+1, paper.Title, paper.Score)
        fmt.Printf("标签: %v\n", paper.Tags)
        fmt.Printf("价值: %s\n", paper.Justification)
        fmt.Printf("HTML: %s\n", paper.HtmlURL)
        fmt.Printf("PDF: %s\n", paper.PdfURL)
        
        // 这里调用你的文章生成和发布逻辑
        // generateAndPublishArticle(paper)
    }
}

func crawlAndAnalysis(service *autogen_service.AutogenService) {
    fmt.Println("=== 爬取并分析论文 ===")
    
    // 先爬取最新论文
    crawler := crawler_service.NewArxivCrawler()
    err := crawler.CrawlPaperAbstract(crawler_service.CS_AI, 50)
    if err != nil {
        log.Printf("爬取失败: %v", err)
        return
    }
    
    // 再分析
    directAnalysis(service)
}
```

## 🔧 配置说明

### AI模型配置

当前使用 `gpt-4o-mini`，每日200次调用限制。

### Redis配置

- **用途**: 仅存储AI分析结果缓存
- **数据库**: Redis DB2
- **缓存键**: `ai_analysis:{ArxivID}`
- **过期时间**: 7天自动清理
- **不存储**: 爬虫论文数据（实时获取）

### 论文来源

支持多个ArXiv分类：
- `CategoryAI`: 人工智能
- `CategoryQuantumPhysics`: 量子物理
- `CategoryAstroPhysics`: 天体物理学
- `CategoryHighEnergyPhysics`: 高能物理实验
- 等等...

### 链接生成规则

- PDF链接格式: `https://arxiv.org/pdf/{ArxivID}.pdf`
- HTML链接格式: `https://arxiv.org/abs/{ArxivID}`

### 缓存管理

```go
// 获取缓存统计
stats, err := service.GetCacheStats()
fmt.Printf("缓存条目: %d\n", stats["total_cached"])

// 清理所有分析缓存
err = service.ClearAnalysisCache()
```

## 📞 技术支持

如有问题，请检查：
1. Redis连接是否正常
2. AI API配置是否正确
3. 论文数据是否存在
4. 日志输出的错误信息
5. ArXiv链接是否可访问

---

*最后更新：2025-01-23* 