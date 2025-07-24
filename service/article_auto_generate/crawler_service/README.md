# ArXiv 多领域爬虫服务

这个爬虫服务支持爬取 ArXiv.org 多个学科领域的最新论文，包括**人工智能、天体物理学、高能物理、量子物理、数学、计算机科学、物理学**等，获取今天发布的**所有论文**（包括新提交、交叉提交、替换提交），并将数据存储到 Redis 中供后续的论文推荐业务使用。

## 功能特性

- 🔍 支持**7个主要学科领域**的论文爬取（154-233篇/领域）
- 📄 获取论文的标题、作者、摘要和链接
- 🔄 支持三种类型：新提交、交叉提交、替换提交
- 🏷️ **自动识别类别**：每篇论文包含中文类别名称
- 🗄️ 数据存储到 Redis 中（可选）
- 🔎 论文搜索功能
- ⚡ 异步获取论文详细摘要
- 🕒 自动控制爬取频率，避免对服务器造成压力
- 🛠️ **String()方法**：类别可直接转换为中文名称

## 支持的学科领域

| 类别 | 中文名称 | ArXiv代码 | URL |
|------|----------|-----------|-----|
| CategoryAI | 人工智能 | cs.AI | https://arxiv.org/list/cs.AI/new |
| CategoryAstroPhysics | 天体物理学 | astro-ph | https://arxiv.org/list/astro-ph/new |
| CategoryHighEnergyPhysics | 高能物理实验 | hep-ex | https://arxiv.org/list/hep-ex/new |
| CategoryQuantumPhysics | 量子物理 | quant-ph | https://arxiv.org/list/quant-ph/new |
| CategoryMathematics | 数学 | math | https://arxiv.org/list/math/new |
| CategoryComputerScience | 计算机科学 | cs | https://arxiv.org/list/cs/new |
| CategoryPhysics | 物理学 | physics | https://arxiv.org/list/physics/new |

## 数据结构

```go
type ArxivPaper struct {
    ArxivID      string        `json:"arxivId"`      // arXiv ID，如 arXiv:2507.16796
    Title        string        `json:"title"`        // 论文标题
    Authors      string        `json:"authors"`      // 作者列表
    Abstract     string        `json:"abstract"`     // 摘要
    PdfURL       string        `json:"pdfUrl"`       // PDF链接
    HtmlURL      string        `json:"htmlUrl"`      // HTML链接
    Category     ArxivCategory `json:"category"`     // 论文类别（枚举）
    CategoryName string        `json:"categoryName"` // 类别中文名称
    CrawlTime    string        `json:"crawlTime"`    // 爬取时间
}

// 类别枚举
type ArxivCategory int
const (
    CategoryAI ArxivCategory = iota + 1
    CategoryAstroPhysics
    CategoryHighEnergyPhysics
    // ... 其他类别
)

// String() 方法示例
func (c ArxivCategory) String() string {
    // 返回中文名称：如 "人工智能"、"天体物理学" 等
}
```

## 基本使用

### 1. 创建爬虫实例

```go
import "blogX_server/service/crawler_service"

// 默认AI类别（向后兼容）
crawler := crawlerservice.NewArxivCrawler()

// 指定特定类别
crawler := crawlerservice.NewArxivCrawlerWithCategory(crawlerservice.CategoryAstroPhysics)
```

### 2. 爬取不同领域的论文

```go
// 方法一：使用通用方法
papers, err := crawler.CrawlRecentPapers()

// 方法二：使用便利方法
papers, err := crawlerservice.CrawlAstrophysicsPapers()        // 天体物理学
papers, err := crawlerservice.CrawlHighEnergyPhysicsPapers()   // 高能物理实验
papers, err := crawlerservice.CrawlQuantumPhysicsPapers()      // 量子物理
papers, err := crawlerservice.CrawlMathematicsPapers()         // 数学
papers, err := crawlerservice.CrawlComputerSciencePapers()     // 计算机科学
papers, err := crawlerservice.CrawlPhysicsPapers()             // 物理学

// 方法三：动态指定类别
papers, err := crawlerservice.CrawlPapersByCategory(crawlerservice.CategoryQuantumPhysics)

fmt.Printf("爬取到 %d 篇论文\n", len(papers)) // 通常154-233篇
```

### 3. 使用类别信息

```go
for _, paper := range papers {
    fmt.Printf("论文: %s\n", paper.Title)
    fmt.Printf("类别: %s (%s)\n", paper.CategoryName, paper.Category.GetCode())
    fmt.Printf("领域: %s\n", paper.Category.String()) // 直接输出中文名称
}
```

### 3. 爬取单篇论文详细信息

```go
paper, err := crawler.CrawlPaperAbstract("arXiv:2507.16796")
if err != nil {
    log.Printf("爬取摘要失败: %v", err)
    return
}

fmt.Printf("摘要: %s\n", paper.Abstract)
```

### 4. 保存到 Redis

```go
err := crawler.SaveToRedis(papers)
if err != nil {
    log.Printf("保存失败: %v", err)
}
```

### 5. 从 Redis 读取

```go
papers, err := crawler.GetFromRedis(20) // 获取20篇论文
if err != nil {
    log.Printf("读取失败: %v", err)
    return
}
```

## 高级功能

### 一键爬取并保存

```go
papers, err := crawler.CrawlAndSave()
if err != nil {
    log.Printf("操作失败: %v", err)
    return
}
// 这个方法会自动爬取论文列表，异步获取前10篇的详细摘要，并保存到Redis
```

### 获取推荐数据

```go
import "blogX_server/service/crawler_service"

papers, err := crawlerservice.GetRecommendationData()
if err != nil {
    log.Printf("获取推荐数据失败: %v", err)
    return
}
// 返回有完整摘要的论文，适合用于推荐算法
```

### 搜索论文

```go
results, err := crawlerservice.SearchPapers("machine learning", 10)
if err != nil {
    log.Printf("搜索失败: %v", err)
    return
}
// 在已爬取的论文中搜索包含指定关键词的论文
```

### 获取 JSON 格式数据

```go
jsonData, err := crawlerservice.GetPapersAsJSON(20)
if err != nil {
    log.Printf("获取JSON失败: %v", err)
    return
}

fmt.Println(jsonData)
```

## Redis 存储说明

- **论文详情**: `arxiv:papers:arXiv:2507.16796`（单个论文的JSON数据）
- **论文列表**: `arxiv:papers:list`（论文ID的有序列表）
- **过期时间**: 24小时自动过期

## 🆕 多领域功能

### 类别管理

```go
// 获取所有支持的类别
categories := crawlerservice.GetAllCategories()
for _, category := range categories {
    fmt.Printf("- %s (%s): %s\n", 
        category.String(),           // 中文名称
        category.GetCode(),          // ArXiv代码
        category.GetEnglishName())   // 英文名称
}

// 根据代码查找类别
category, err := crawlerservice.GetCategoryByCode("astro-ph")
if err == nil {
    fmt.Printf("找到类别: %s\n", category.String()) // 输出：天体物理学
}
```

### 批量爬取多个领域

```go
// 爬取多个感兴趣的领域
categories := []crawlerservice.ArxivCategory{
    crawlerservice.CategoryAI,
    crawlerservice.CategoryQuantumPhysics,
    crawlerservice.CategoryAstroPhysics,
}

allPapers := make([]crawlerservice.ArxivPaper, 0)
for _, category := range categories {
    papers, err := crawlerservice.CrawlPapersByCategory(category)
    if err != nil {
        log.Printf("爬取 %s 失败: %v", category.String(), err)
        continue
    }
    
    log.Printf("✅ 成功爬取 %s 论文 %d 篇", category.String(), len(papers))
    allPapers = append(allPapers, papers...)
}

log.Printf("总共爬取论文 %d 篇", len(allPapers))
```

### 论文推荐数据生成

```go
// 为不同领域生成推荐标签
func generateRecommendationTags(papers []crawlerservice.ArxivPaper) map[string][]crawlerservice.ArxivPaper {
    taggedPapers := make(map[string][]crawlerservice.ArxivPaper)
    
    for _, paper := range papers {
        categoryName := paper.Category.String() // 自动获取中文名称
        taggedPapers[categoryName] = append(taggedPapers[categoryName], paper)
    }
    
    return taggedPapers
}

// 使用示例
papers, _ := crawlerservice.CrawlPapersByCategory(crawlerservice.CategoryAI)
tags := generateRecommendationTags(papers)

for tag, tagPapers := range tags {
    fmt.Printf("📚 %s 领域论文: %d 篇\n", tag, len(tagPapers))
}
```

## 测试

```go
import "blogX_server/service/crawler_service"

// 运行完整测试
crawlerservice.TestArxivCrawler()
```

## 注意事项

1. **请求频率**: 自动控制请求频率，避免对 ArXiv 服务器造成压力
2. **错误处理**: 网络请求可能失败，已包含重试和错误处理逻辑
3. **Redis 可选**: 如果 Redis 未连接，爬虫仍然可以正常工作，只是不会持久化数据
4. **异步处理**: 详细摘要的爬取是异步进行的，可能需要等待一段时间

## 未来扩展

- 支持更多 ArXiv 分类（如 cs.CV, cs.LG 等）
- 添加论文引用关系爬取
- 支持按时间范围爬取历史论文
- 添加论文相似度计算
- 集成更多推荐算法 