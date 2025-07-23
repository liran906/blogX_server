package crawlerservice

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// DebugArxivHTML 调试ArXiv页面的HTML结构
func DebugArxivHTML() error {
	logrus.Info("开始调试ArXiv页面结构...")

	// 创建HTTP请求 - 默认使用AI类别进行调试
	req, err := http.NewRequest("GET", CategoryAI.GetURL(), nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	logrus.Infof("HTTP状态码: %d", resp.StatusCode)
	logrus.Infof("响应头: %v", resp.Header)

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("解析HTML失败: %v", err)
	}

	// 获取页面标题
	title := doc.Find("title").Text()
	logrus.Infof("页面标题: %s", title)

	// 保存HTML到文件用于分析
	htmlContent, err := doc.Html()
	if err == nil {
		err = os.WriteFile("/tmp/arxiv_debug.html", []byte(htmlContent), 0644)
		if err == nil {
			logrus.Info("HTML内容已保存到 /tmp/arxiv_debug.html")
		}
	}

	// 分析可能的论文容器
	logrus.Info("=== 分析页面结构 ===")

	// 检查常见的容器
	containers := []string{"dd", "dl", ".list-title", ".abs", "dt", "li"}
	for _, container := range containers {
		count := doc.Find(container).Length()
		logrus.Infof("发现 %d 个 '%s' 元素", count, container)

		if count > 0 && count < 20 { // 如果数量合理，输出前几个的内容
			doc.Find(container).Each(func(i int, s *goquery.Selection) {
				if i < 3 { // 只输出前3个
					text := s.Text()
					if len(text) > 100 {
						text = text[:100] + "..."
					}
					logrus.Infof("  [%d] %s: %s", i, container, text)
				}
			})
		}
	}

	// 检查是否有论文标识符链接
	logrus.Info("=== 检查arXiv ID链接 ===")
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && (contains(href, "/abs/") || contains(href, "arxiv")) {
			if i < 5 { // 只输出前5个
				logrus.Infof("发现arXiv链接[%d]: %s -> %s", i, s.Text(), href)
			}
		}
	})

	// 检查包含"Title:"的元素
	logrus.Info("=== 检查标题元素 ===")
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if contains(text, "Title:") && len(text) < 200 {
			className, _ := s.Attr("class")
			tagName := goquery.NodeName(s)
			logrus.Infof("发现标题元素: <%s class='%s'>%s</%s>", tagName, className, text, tagName)
		}
	})

	logrus.Info("调试完成，请检查输出信息")
	return nil
}

// contains 简单的字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		s != substr &&
		findSubstring(s, substr) >= 0
}

// findSubstring 查找子字符串位置
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
