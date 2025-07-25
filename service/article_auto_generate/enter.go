// Path: ./service/article_auto_generate/enter.go

package article_auto_generate

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	"blogX_server/service/article_auto_generate/batch_scoring_service"
	"blogX_server/service/article_auto_generate/crawler_service"
	"blogX_server/service/common_utils"
	"blogX_server/service/email_service"
	"blogX_server/service/redis_service/redis_ai_cache"
	"blogX_server/utils/markdown"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

func AutoGenerateArxivAbstract() {
	limit := global.Config.Site.AutoGen.Limit
	top := global.Config.Site.AutoGen.Top
	for _, code := range global.Config.Site.AutoGen.Categories {
		c, err := crawler_service.GetCategoryByCode(code)
		if err != nil {
			logrus.Errorf("get arxiv category error: %v", err)
			continue
		}
		autoGenerateArxivAbstract(c, limit, top)
	}
}

func autoGenerateArxivAbstract(category crawler_service.ArxivCategory, limit, topN int) {
	logrus.Infof("开始自动生成 %s 类别的论文摘要，限制：%d篇，选择Top：%d", category.String(), limit, topN)

	// 1. 指定类别爬取论文
	crawler := crawler_service.NewArxivCrawlerWithCategory(category)
	papers, err := crawler.CrawlRecentPapers()
	if err != nil {
		logrus.Errorf("爬取论文失败: %v", err)
		return
	}

	if len(papers) == 0 {
		logrus.Warnf("未爬取到 %s 类别的论文", category.String())
		return
	}

	// 限制爬取数量
	if limit > 0 && limit < len(papers) {
		papers = papers[:limit]
	}

	logrus.Infof("成功爬取 %s 类别论文 %d 篇", category.String(), len(papers))

	// 2. 使用新的批次评分系统进行两阶段分析
	config := batch_scoring_service.DefaultBatchScoringConfig()
	config.TopN = global.Config.Site.AutoGen.Top // 使用配置文件中的Top值

	analyzer := batch_scoring_service.NewTwoStageAnalyzer(config)

	// 执行两阶段分析
	request := batch_scoring_service.TwoStageAnalysisRequest{
		Papers: papers,
	}

	result, err := analyzer.AnalyzeTwoStage(request)
	if err != nil {
		logrus.Errorf("两阶段分析失败: %v", err)
		return
	}

	logrus.Infof("两阶段分析完成：第一阶段评分%d篇，第二阶段详细分析%d篇，平均分%.1f",
		len(result.Stage1Results), len(result.Stage2Results), result.Statistics.AverageScore)

	// 3. 异步保存分析结果到缓存
	go func() {
		err := redis_ai_cache.SaveBatchScoringResult(result)
		if err != nil {
			logrus.Errorf("保存分析结果缓存失败: %v", err)
		}
	}()

	// 4. 格式化分析报告
	content := formatTwoStageAnalysisReport(result, category.String(), papers)

	// 5. 生成文章
	err = articleGen(content, category.String())
	if err != nil {
		logrus.Errorf("文章自动生成失败: %v", err)
		return
	}

	logrus.Infof("%s 类别论文摘要自动生成完成", category.String())
}

// formatTwoStageAnalysisReport 格式化两阶段分析报告
func formatTwoStageAnalysisReport(result *batch_scoring_service.TwoStageAnalysisResult, categoryName string, originalPapers []crawler_service.ArxivPaper) string {
	var content string

	// 标题和统计概要
	content += fmt.Sprintf("# %s 领域论文智能分析报告\n\n", categoryName)
	content += fmt.Sprintf("🎯 **分析概要**\n")
	content += fmt.Sprintf("- 📊 总论文数：%d 篇\n", result.Statistics.TotalPapers)
	content += fmt.Sprintf("- ⭐ 平均评分：%.1f 分\n", result.Statistics.AverageScore)
	content += fmt.Sprintf("- 🏆 最高评分：%.1f 分\n", result.Statistics.MaxScore)
	content += fmt.Sprintf("- 📈 详细分析：%d 篇高质量论文\n\n", result.Statistics.Stage2SelectedCount)

	// 分数分布统计
	content += "📈 **评分分布**\n"
	for scoreRange, count := range result.Statistics.ScoreDistribution {
		content += fmt.Sprintf("- %s分：%d 篇\n", scoreRange, count)
	}
	content += "\n"

	// 高质量论文详细分析
	if len(result.Stage2Results) > 0 {
		content += "## 🏆 高质量论文详细分析\n\n"

		// 按照第一阶段的分数排序第二阶段结果
		sortedStage2 := make([]batch_scoring_service.DetailedAnalysis, len(result.Stage2Results))
		copy(sortedStage2, result.Stage2Results)

		// 建立ArxivID到分数的映射
		scoreMap := make(map[string]float64)
		for _, paper := range result.Stage1Results {
			scoreMap[paper.ArxivID] = paper.FinalScore
		}

		// 按分数排序
		sort.Slice(sortedStage2, func(i, j int) bool {
			score1 := scoreMap[sortedStage2[i].ArxivID]
			score2 := scoreMap[sortedStage2[j].ArxivID]
			return score1 > score2
		})

		for i, analysis := range sortedStage2 {
			score := scoreMap[analysis.ArxivID]
			scoreEmoji := common_utils.GetScoreEmoji(score)

			content += fmt.Sprintf("### %d. %s\n\n", i+1, common_utils.TruncateString(analysis.Title, 120))
			content += fmt.Sprintf("**作者**: %s\n\n", common_utils.TruncateString(analysis.Authors, 200))
			content += fmt.Sprintf("**AI评分**: %s %.1f分\n\n", scoreEmoji, score)

			// 添加论文源链接
			htmlURL := fmt.Sprintf("https://arxiv.org/abs/%s", analysis.ArxivID)
			pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", analysis.ArxivID)
			content += fmt.Sprintf("**论文源**: [`%s`](%s) | [`PDF`](%s)\n\n", analysis.ArxivID, htmlURL, pdfURL)

			// 关键词
			if len(analysis.Tags) > 0 {
				content += "**关键词**: "
				for j, tag := range analysis.Tags {
					if j > 0 {
						content += " | "
					}
					content += fmt.Sprintf("`%s`", tag)
				}
				content += "\n\n"
			}

			// 中文摘要
			content += fmt.Sprintf("**中文摘要**\n%s\n\n", analysis.Summary)

			// 专业评价
			content += fmt.Sprintf("**AI评价**\n%s\n\n", analysis.Evaluation)

			content += "---\n\n"
		}
	}

	// 剩余论文概览（显示未进行详细分析的论文）
	content += "## 📊 其他论文概览\n\n"

	// 按分数排序第一阶段结果
	sortedStage1 := make([]batch_scoring_service.PaperScore, len(result.Stage1Results))
	copy(sortedStage1, result.Stage1Results)
	sort.Slice(sortedStage1, func(i, j int) bool {
		return sortedStage1[i].FinalScore > sortedStage1[j].FinalScore
	})

	// 创建详细分析论文的ArxivID集合
	detailedAnalysisSet := make(map[string]bool)
	for _, analysis := range result.Stage2Results {
		detailedAnalysisSet[analysis.ArxivID] = true
	}

	// 显示不在详细分析中的剩余论文
	count := len(result.Stage2Results) + 1
	for _, paper := range sortedStage1 {
		// 跳过已经进行详细分析的论文
		if detailedAnalysisSet[paper.ArxivID] {
			continue
		}

		scoreEmoji := common_utils.GetScoreEmoji(paper.FinalScore)

		// 生成链接
		htmlURL := fmt.Sprintf("https://arxiv.org/abs/%s", paper.ArxivID)
		pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", paper.ArxivID)

		// 获取真实标题
		title := paper.ArxivID // 默认使用ArxivID
		for _, originalPaper := range originalPapers {
			if originalPaper.ArxivID == paper.ArxivID {
				title = originalPaper.Title
				break
			}
		}

		content += fmt.Sprintf("%d. [%s](%s) | [PDF](%s)\n",
			count, common_utils.TruncateString(title, 120), htmlURL, pdfURL)

		// 显示分项评分
		if paper.Score1 != nil {
			content += fmt.Sprintf("   - AI评分: %s **%.1f分** | 创新性: %d | 技术深度: %d | 实用性: %d\n",
				scoreEmoji, paper.FinalScore, paper.Score1.Innovation, paper.Score1.Technical, paper.Score1.Practical)
		}

		count++
	}

	content += "\n\n---\n\n"
	content += fmt.Sprintf("*本报告由AI智能分析系统自动生成，不代表平台观点，分析时间：%s*\n", time.Now().Format("2006-01-02 15:04:05"))

	return content
}

// 注意：工具函数已迁移到 service/common_utils 包中

func articleGen(content, category string) error {
	uid := global.Config.Site.AutoGen.UserID

	// 取分类
	var cat models.CategoryModel
	if category != "" {
		err := global.DB.Take(&cat, "name = ? and user_id = ?", category, uid).Error
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				cat.Name = category
				cat.UserID = uid
				err = global.DB.Create(&cat).Error
				if err != nil {
					logrus.Errorf("创建新分类失败 %v", err)
				}
			} else {
				logrus.Errorf("文章分类错误 %v", err)
			}
		}
	}

	nowDate := time.Now().Format("01月02日")
	abstract := fmt.Sprintf("每天让 AI 助手，从特定领域 百篇最新论文 中智能评分筛选最好的论文，供您阅读")

	var article = models.ArticleModel{
		Title:          nowDate + " " + category + " 智能分析报告",
		Abstract:       abstract,
		CoverURL:       "/uploads/images/9742aaccce6aaf3078e1f9df8bcc222d.png",
		Content:        content,
		CategoryID:     &cat.ID,
		Tags:           ctype.List{category, "AI分析", "智能评分"},
		OpenForComment: true,
		UserID:         uid,
		Status:         enum.ArticleStatusPublish, // 自动免审
	}

	// 入库
	err := global.DB.Create(&article).Error
	if err != nil {
		logrus.Errorf("文章自动发布失败，%v", err)
		return err
	}
	logrus.Info("文章自动生成发布成功")

	sendToSubscribers(&article, category)
	return nil
}

func sendToSubscribers(article *models.ArticleModel, category string) {
	content := injectLink(article.Content, article.ID)
	html := markdown.MdToHTML(content)
	var emails []string
	var subs []models.UserConfigModel
	err := global.DB.Preload("UserModel").Where("subscribe = ?", true).Find(&subs).Error
	if err != nil {
		logrus.Errorf("获取订阅用户失败: %v", err)
		return
	}
	for _, sub := range subs {
		emails = append(emails, sub.UserModel.Email)
	}
	err = email_service.SendSubscribe(emails, category, html)
	if err != nil {
		logrus.Errorf("订阅邮件发送失败: %v", err)
	}
	logrus.Info("订阅邮件发送成功")
}

func TestFunc(content, category string) {
	to := []string{"liran900620@gmail.com"}
	email_service.SendSubscribe(to, category, content)
}

func injectLink(md string, articleID uint) string {
	lines := strings.SplitN(md, "\n", 2) // 只拆前两部分
	if len(lines) == 0 {
		return md
	}
	date := time.Now().Format("01月02日 ")
	link := fmt.Sprintf("[[原文]](https://blog.golir.top/article/%d)\n", articleID)

	header := "# " + date + string([]byte(lines[0])[1:]) + link

	return header + lines[1]
}
