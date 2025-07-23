package crawlerservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"blogX_server/global"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

const (
	ArxivBaseURL = "https://arxiv.org"
	UserAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// Redis键前缀
	RedisKeyPrefix = "arxiv:papers:"
	RedisListKey   = "arxiv:papers:list"
)

// ArxivPaper 简化的ArXiv论文结构体
type ArxivPaper struct {
	ArxivID       string        `json:"arxivId"`       // arXiv ID，如 arXiv:2507.16796
	Title         string        `json:"title"`         // 论文标题
	Authors       string        `json:"authors"`       // 作者列表
	Abstract      string        `json:"abstract"`      // 摘要
	PdfURL        string        `json:"pdfUrl"`        // PDF链接
	HtmlURL       string        `json:"htmlUrl"`       // HTML链接
	Category      ArxivCategory `json:"category"`      // 论文类别
	CategoryName  string        `json:"categoryName"`  // 类别中文名称
	PublishedDate string        `json:"publishedDate"` // 发表时间
	CrawlTime     string        `json:"crawlTime"`     // 爬取时间
}

// ArxivCrawler ArXiv爬虫结构体
type ArxivCrawler struct {
	client   *http.Client
	category ArxivCategory
}

// NewArxivCrawler 创建新的ArXiv爬虫实例（默认AI类别）
func NewArxivCrawler() *ArxivCrawler {
	return NewArxivCrawlerWithCategory(CategoryAI)
}

// NewArxivCrawlerWithCategory 创建指定类别的ArXiv爬虫实例
func NewArxivCrawlerWithCategory(category ArxivCategory) *ArxivCrawler {
	return &ArxivCrawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		category: category,
	}
}

// CrawlRecentPapers 爬取指定类别的最新论文
func (ac *ArxivCrawler) CrawlRecentPapers() ([]ArxivPaper, error) {
	logrus.Infof("开始爬取ArXiv %s论文...", ac.category.String())

	// 创建HTTP请求
	req, err := http.NewRequest("GET", ac.category.GetURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// 发送请求
	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 解析论文条目
	papers, err := ac.parseArxivPapers(doc)
	if err != nil {
		return nil, fmt.Errorf("解析论文失败: %v", err)
	}

	logrus.Infof("成功爬取 %d 篇论文", len(papers))
	return papers, nil
}

// parseArxivPapers 解析论文信息
func (ac *ArxivCrawler) parseArxivPapers(doc *goquery.Document) ([]ArxivPaper, error) {
	var papers []ArxivPaper
	crawlTime := time.Now().Format("2006-01-02 15:04:05")

	// ArXiv页面结构：<dt>包含arXiv ID链接，紧接着的<dd>包含论文详情
	doc.Find("dt").Each(func(i int, dtElement *goquery.Selection) {
		paper := ArxivPaper{
			Category:     ac.category,
			CategoryName: ac.category.String(),
			CrawlTime:    crawlTime,
		}

		// 1. 从<dt>中解析ArXiv ID和链接
		if link := dtElement.Find("a[href*='/abs/']").First(); link.Length() > 0 {
			href, exists := link.Attr("href")
			if exists {
				// 从链接中提取ArXiv ID，如 /abs/2507.16796
				re := regexp.MustCompile(`/abs/(\d+\.\d+)`)
				matches := re.FindStringSubmatch(href)
				if len(matches) > 1 {
					paper.ArxivID = "arXiv:" + matches[1]
					paper.PdfURL = ArxivBaseURL + "/pdf/" + matches[1] + ".pdf"
					paper.HtmlURL = ArxivBaseURL + "/html/" + matches[1]
				}
			}
		}

		// 2. 查找对应的<dd>元素（紧跟在<dt>后面）
		ddElement := dtElement.Next()
		if ddElement.Length() == 0 || ddElement.Get(0).Data != "dd" {
			return // 没有找到对应的dd元素
		}

		// 检查是否是论文条目
		if ddElement.Find("div[class*='list-title']").Length() == 0 {
			return
		}

		// 3. 从<dd>中解析标题
		if title := ddElement.Find("div[class*='list-title']"); title.Length() > 0 {
			titleText := title.Text()
			titleText = strings.TrimPrefix(titleText, "Title:")
			paper.Title = strings.TrimSpace(titleText)
		}

		// 4. 从<dd>中解析作者
		if authors := ddElement.Find("div[class*='list-authors']"); authors.Length() > 0 {
			authorsText := authors.Text()
			authorsText = strings.TrimPrefix(authorsText, "Authors:")
			paper.Authors = strings.TrimSpace(authorsText)
		}

		// 5. 从<dd>中解析摘要 - ArXiv "new"页面包含完整摘要！
		if abstract := ddElement.Find("p.mathjax"); abstract.Length() > 0 {
			abstractText := abstract.Text()
			abstractText = strings.TrimSpace(abstractText)
			paper.Abstract = abstractText
		}

		// 6. 从<dd>中解析评论信息（可选）
		var comments string
		if commentDiv := ddElement.Find("div[class*='list-comments']"); commentDiv.Length() > 0 {
			commentsText := commentDiv.Text()
			commentsText = strings.TrimPrefix(commentsText, "Comments:")
			comments = strings.TrimSpace(commentsText)
		}

		// 7. 从<dd>中解析学科分类
		var subjects string
		if subjectDiv := ddElement.Find("div[class*='list-subjects']"); subjectDiv.Length() > 0 {
			subjectsText := subjectDiv.Text()
			subjectsText = strings.TrimPrefix(subjectsText, "Subjects:")
			subjects = strings.TrimSpace(subjectsText)
		}

		// 8. 从<dd>中解析发表时间 - 通常在文档的开头或者可以从ArXiv ID推断
		paper.PublishedDate = extractPublishedDateFromArxivID(paper.ArxivID)

		// 9. 只添加有有效ArXiv ID和标题的论文
		if paper.ArxivID != "" && paper.Title != "" {
			papers = append(papers, paper)
			logrus.Infof("解析论文: %s - %s (摘要长度: %d)", paper.ArxivID, paper.Title, len(paper.Abstract))

			// 如果有评论或学科信息，可以记录到日志中
			if comments != "" {
				logrus.Debugf("论文 %s 评论: %s", paper.ArxivID, comments)
			}
			if subjects != "" {
				logrus.Debugf("论文 %s 学科: %s", paper.ArxivID, subjects)
			}
		}
	})

	logrus.Infof("共解析到 %d 篇论文", len(papers))
	return papers, nil
}

// CrawlPaperDetails 爬取单篇论文的详细信息（摘要已在列表页面获取）
func (ac *ArxivCrawler) CrawlPaperDetails(arxivID string) (*ArxivPaper, error) {
	// 构建论文详情页URL
	cleanID := strings.TrimPrefix(arxivID, "arXiv:")
	detailURL := fmt.Sprintf("%s/abs/%s", ArxivBaseURL, cleanID)

	logrus.Infof("爬取论文详细信息: %s", detailURL)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	// 发送请求
	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	paper := &ArxivPaper{
		ArxivID:      arxivID,
		Category:     ac.category,
		CategoryName: ac.category.String(),
		CrawlTime:    time.Now().Format("2006-01-02 15:04:05"),
	}

	// 解析标题（更详细版本）
	if title := doc.Find("h1.title"); title.Length() > 0 {
		titleText := title.Text()
		titleText = strings.TrimPrefix(titleText, "Title:")
		paper.Title = strings.TrimSpace(titleText)
	}

	// 解析作者（更详细版本，包含机构信息）
	if authors := doc.Find(".authors"); authors.Length() > 0 {
		authorsText := authors.Text()
		authorsText = strings.TrimPrefix(authorsText, "Authors:")
		paper.Authors = strings.TrimSpace(authorsText)
	}

	// 解析摘要（详情页面可能有更完整的格式）
	if abstract := doc.Find(".abstract"); abstract.Length() > 0 {
		abstractText := abstract.Text()
		abstractText = strings.TrimPrefix(abstractText, "Abstract:")
		paper.Abstract = strings.TrimSpace(abstractText)
	}

	// 解析更多详细信息
	var additionalInfo []string

	// 解析DOI信息
	if doi := doc.Find("td.doi"); doi.Length() > 0 {
		doiText := doi.Text()
		if doiText != "" {
			additionalInfo = append(additionalInfo, fmt.Sprintf("DOI: %s", strings.TrimSpace(doiText)))
		}
	}

	// 解析期刊引用信息
	if journal := doc.Find("td.jref"); journal.Length() > 0 {
		journalText := journal.Text()
		if journalText != "" {
			additionalInfo = append(additionalInfo, fmt.Sprintf("期刊: %s", strings.TrimSpace(journalText)))
		}
	}

	// 解析提交历史
	if history := doc.Find(".submission-history"); history.Length() > 0 {
		historyText := history.Text()
		if historyText != "" {
			additionalInfo = append(additionalInfo, fmt.Sprintf("提交历史: %s", strings.TrimSpace(historyText)))
		}
	}

	// 如果有额外信息，添加到摘要末尾或专门的字段
	if len(additionalInfo) > 0 {
		logrus.Debugf("论文 %s 的额外信息: %v", arxivID, additionalInfo)
	}

	// 设置链接
	cleanID = strings.TrimPrefix(arxivID, "arXiv:")
	paper.PdfURL = ArxivBaseURL + "/pdf/" + cleanID + ".pdf"
	paper.HtmlURL = ArxivBaseURL + "/html/" + cleanID

	logrus.Infof("成功爬取论文详细信息: %s", arxivID)
	return paper, nil
}

// SaveToRedis 保存论文列表到Redis
func (ac *ArxivCrawler) SaveToRedis(papers []ArxivPaper) error {
	if global.Redis == nil {
		logrus.Warn("Redis未连接，跳过保存")
		return fmt.Errorf("redis未连接")
	}

	if len(papers) == 0 {
		logrus.Warn("没有论文需要保存到Redis")
		return nil
	}

	// 清除旧的列表
	err := global.Redis.Del(RedisListKey).Err()
	if err != nil {
		logrus.Errorf("清除Redis列表失败: %v", err)
	}

	successCount := 0

	for _, paper := range papers {
		// 将论文转换为JSON
		paperJSON, err := json.Marshal(paper)
		if err != nil {
			logrus.Errorf("序列化论文失败 %s: %v", paper.ArxivID, err)
			continue
		}

		// 保存单个论文
		redisKey := RedisKeyPrefix + paper.ArxivID
		err = global.Redis.Set(redisKey, paperJSON, 24*time.Hour).Err() // 24小时过期
		if err != nil {
			logrus.Errorf("保存论文到Redis失败 %s: %v", paper.ArxivID, err)
			continue
		}

		// 添加到列表
		err = global.Redis.LPush(RedisListKey, paper.ArxivID).Err()
		if err != nil {
			logrus.Errorf("添加到Redis列表失败 %s: %v", paper.ArxivID, err)
			continue
		}

		successCount++
	}

	// 设置列表过期时间
	global.Redis.Expire(RedisListKey, 24*time.Hour)

	logrus.Infof("成功保存 %d 篇论文到Redis", successCount)
	return nil
}

// GetFromRedis 从Redis获取论文列表
func (ac *ArxivCrawler) GetFromRedis(limit int) ([]ArxivPaper, error) {
	if global.Redis == nil {
		return nil, fmt.Errorf("redis未连接")
	}

	// 获取论文ID列表
	arxivIDs, err := global.Redis.LRange(RedisListKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis列表失败: %v", err)
	}

	var papers []ArxivPaper

	for _, arxivID := range arxivIDs {
		redisKey := RedisKeyPrefix + arxivID
		paperJSON, err := global.Redis.Get(redisKey).Result()
		if err != nil {
			logrus.Errorf("获取论文失败 %s: %v", arxivID, err)
			continue
		}

		var paper ArxivPaper
		err = json.Unmarshal([]byte(paperJSON), &paper)
		if err != nil {
			logrus.Errorf("反序列化论文失败 %s: %v", arxivID, err)
			continue
		}

		papers = append(papers, paper)
	}

	logrus.Infof("从Redis获取到 %d 篇论文", len(papers))
	return papers, nil
}

// CrawlAndSave 爬取论文并保存到Redis
func (ac *ArxivCrawler) CrawlAndSave() ([]ArxivPaper, error) {
	// 爬取论文列表
	papers, err := ac.CrawlRecentAIPapers()
	if err != nil {
		return nil, err
	}

	// 为前几篇论文获取额外详细信息（异步处理）
	go func() {
		maxDetails := 5 // 最多详细爬取5篇，因为摘要已经有了
		if len(papers) < maxDetails {
			maxDetails = len(papers)
		}

		for i := 0; i < maxDetails; i++ {
			// 获取额外的详细信息，如DOI、详细分类等
			detailedPaper, err := ac.CrawlPaperDetails(papers[i].ArxivID)
			if err != nil {
				logrus.Errorf("爬取论文详情失败 %s: %v", papers[i].ArxivID, err)
				continue
			}

			// 如果详情页面有更完整的摘要，则更新
			if len(detailedPaper.Abstract) > len(papers[i].Abstract) {
				papers[i].Abstract = detailedPaper.Abstract
			}

			logrus.Infof("已获取论文 %s 的详细信息", papers[i].ArxivID)
			// 控制频率，避免过快请求
			time.Sleep(2 * time.Second) // 增加间隔，因为是额外请求
		}

		// 保存更新后的论文到Redis
		if err := ac.SaveToRedis(papers); err != nil {
			logrus.Errorf("保存论文到Redis失败: %v", err)
		}
	}()

	// 先保存基本信息到Redis
	if err := ac.SaveToRedis(papers); err != nil {
		logrus.Errorf("保存论文到Redis失败: %v", err)
	}

	return papers, nil
}

// CrawlRecentAIPapers 爬取最近的AI论文（向后兼容）
func (ac *ArxivCrawler) CrawlRecentAIPapers() ([]ArxivPaper, error) {
	// 临时设置为AI类别以保持向后兼容
	originalCategory := ac.category
	ac.category = CategoryAI
	defer func() { ac.category = originalCategory }()

	return ac.CrawlRecentPapers()
}

// 便利方法：直接爬取不同类别的论文

// CrawlAstrophysicsPapers 爬取天体物理学论文
func CrawlAstrophysicsPapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryAstroPhysics)
	return crawler.CrawlRecentPapers()
}

// CrawlHighEnergyPhysicsPapers 爬取高能物理实验论文
func CrawlHighEnergyPhysicsPapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryHighEnergyPhysics)
	return crawler.CrawlRecentPapers()
}

// CrawlQuantumPhysicsPapers 爬取量子物理论文
func CrawlQuantumPhysicsPapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryQuantumPhysics)
	return crawler.CrawlRecentPapers()
}

// CrawlMathematicsPapers 爬取数学论文
func CrawlMathematicsPapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryMathematics)
	return crawler.CrawlRecentPapers()
}

// CrawlComputerSciencePapers 爬取计算机科学论文
func CrawlComputerSciencePapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryComputerScience)
	return crawler.CrawlRecentPapers()
}

// CrawlPhysicsPapers 爬取物理学论文
func CrawlPhysicsPapers() ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(CategoryPhysics)
	return crawler.CrawlRecentPapers()
}

// CrawlPapersByCategory 根据类别爬取论文
func CrawlPapersByCategory(category ArxivCategory) ([]ArxivPaper, error) {
	crawler := NewArxivCrawlerWithCategory(category)
	return crawler.CrawlRecentPapers()
}

// extractPublishedDateFromArxivID 从ArXiv ID中提取发表时间
// ArXiv ID格式: arXiv:YYMM.NNNNN 例如 arXiv:2507.16796 表示2025年7月
func extractPublishedDateFromArxivID(arxivID string) string {
	if arxivID == "" {
		return ""
	}

	// 移除 "arXiv:" 前缀
	idPart := strings.TrimPrefix(arxivID, "arXiv:")

	// 提取年月部分 (YYMM)
	if len(idPart) < 4 {
		return ""
	}

	yearMonth := idPart[:4]
	if len(yearMonth) != 4 {
		return ""
	}

	// 解析年份和月份
	year := "20" + yearMonth[:2] // YY -> 20YY
	month := yearMonth[2:4]      // MM

	// 验证月份有效性
	monthNum := 0
	if _, err := fmt.Sscanf(month, "%d", &monthNum); err != nil || monthNum < 1 || monthNum > 12 {
		return ""
	}

	// 返回格式化的日期 (YYYY-MM格式)
	return fmt.Sprintf("%s-%s", year, month)
}
