package apis

import (
	"io"
	"net/http"
	"regexp"
	"smart-api/app/article/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ArticleHandler 处理文章上传和内容处理
type ArticleHandler struct {
	api.Api
}

// FrontMatter 定义Markdown文件的Front Matter结构
type FrontMatter struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Author      string   `json:"author"`
	Date        string   `json:"date"`
}

// ArticleResponse 定义文章上传响应结构
type ArticleResponse struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// 定义前端请求的数据结构
type ArticlePublishRequest struct {
	Title       string `json:"title" form:"title"`
	Content     string `json:"content" form:"content"`
	HTMLContent string `json:"html_content" form:"html_content"`
}

// UploadMarkdownFile 处理Markdown文件上传和文章发布
// @Summary 上传Markdown文件并发布文章
// @Description 上传Markdown文件或直接提交内容并发布技术博客文章
// @Tags 文章管理
// @Accept multipart/form-data,application/json
// @Produce json
// @Param request body ArticlePublishRequest false "文章发布请求"
// @Param title formData string false "文章标题"
// @Param content formData string false "Markdown内容"
// @Param file formData file false ".md文件"
// @Success 200 {object} ArticleResponse "{\"code\": 0, \"data\": {\"id\": 1, \"title\": \"文章标题\"}}"
// @Failure 400 {object} response.Response "{\"code\": 400, \"message\": \"错误信息\"}"
// @Router /api/v1/article [post]
// @Security Bearer
func (e *ArticleHandler) UploadMarkdownFile(c *gin.Context) {
	var markdownContent string
	var frontMatter FrontMatter
	var title string // 定义标题变量

	// 确保初始化上下文和ORM
	if err := e.MakeContext(c).MakeOrm().Errors; err != nil {
		e.Logger.Error("初始化失败:", err)
		e.Error(http.StatusInternalServerError, err, "系统初始化失败")
		return
	}

	// 检查请求类型
	contentType := c.Request.Header.Get("Content-Type")
	e.Logger.Debug("请求内容类型:", contentType)

	// 首先尝试解析JSON请求体
	var req ArticlePublishRequest
	e.Logger.Debug("尝试解析JSON请求体")
	if err := c.ShouldBindJSON(&req); err == nil {
		// 成功解析JSON请求体
		e.Logger.Debug("成功解析JSON请求体")
		e.Logger.Debug("JSON请求中的标题:", req.Title)
		e.Logger.Debug("JSON请求中的内容长度:", len(req.Content))
		markdownContent = req.Content
		title = req.Title // 从JSON中获取标题
	} else {
		e.Logger.Debug("JSON解析失败，错误:", err)
		// 如果JSON解析失败，尝试获取表单中的标题和内容
		// 先获取标题
		title = c.PostForm("title")
		e.Logger.Debug("通过PostForm获取到的标题:", title)

		// 尝试多种方式获取文章内容，确保能正确读取前端提交的数据
		// 1. 首先尝试通过PostForm获取
		markdownContent = c.PostForm("content")
		e.Logger.Debug("通过PostForm获取到的内容长度:", len(markdownContent))

		// 2. 如果PostForm获取失败，尝试直接从表单中获取
		if markdownContent == "" {
			if c.Request.Form != nil && c.Request.Form.Has("content") {
				markdownContent = c.Request.Form.Get("content")
				e.Logger.Debug("通过Request.Form获取到的内容长度:", len(markdownContent))
			}
		}

		// 3. 如果表单获取失败，尝试获取多部分表单数据
		if markdownContent == "" {
			// 尝试解析多部分表单
			if err := c.Request.ParseMultipartForm(10 << 20); err == nil {
				e.Logger.Debug("成功解析多部分表单")
				if c.Request.MultipartForm != nil && c.Request.MultipartForm.Value != nil {
					if values, ok := c.Request.MultipartForm.Value["content"]; ok && len(values) > 0 {
						markdownContent = values[0]
						e.Logger.Debug("通过MultipartForm获取到的内容长度:", len(markdownContent))
					}
				}
			} else {
				e.Logger.Debug("解析多部分表单失败:", err)
			}
		}

		// 如果表单中没有内容，再尝试从文件获取
		// 注意：文件上传只是辅助功能，不是发布文章的必要条件
		if markdownContent == "" {
			// 尝试获取文件，但不将错误作为失败原因
			file, header, err := c.Request.FormFile("file")
			if err == nil {
				defer file.Close()

				// 验证文件类型
				if !strings.HasSuffix(strings.ToLower(header.Filename), ".md") {
					e.Error(http.StatusBadRequest, nil, "文件类型错误，请上传.md格式的文件")
					return
				}

				// 验证文件大小（50MB）
				if header.Size > 50*1024*1024 {
					e.Error(http.StatusBadRequest, nil, "文件大小超过限制，最大支持50MB")
					return
				}

				// 读取文件内容
				fileBytes, err := io.ReadAll(file)
				if err != nil {
					e.Logger.Error("读取文件失败:", err)
					// 只记录日志，不返回错误，因为文件上传是辅助功能
				} else {
					markdownContent = string(fileBytes)
					e.Logger.Debug("通过文件获取到的内容长度:", len(markdownContent))
				}
			} else {
				e.Logger.Debug("获取文件失败:", err)
			}
			// 文件获取失败时，我们不返回错误，因为文件上传只是辅助功能
		}
	}

	// 验证文章内容是否存在
	if markdownContent == "" {
		// 记录详细信息以便调试
		e.Logger.Error("未找到文章内容，请求内容类型:", contentType)
		e.Logger.Error("标题是否为空:", title == "")
		e.Logger.Error("请求方法:", c.Request.Method)
		e.Logger.Error("请求路径:", c.Request.URL.Path)
		e.Error(http.StatusBadRequest, nil, "请提供文章内容")
		return
	}

	// 解析Front Matter（如果存在）
	frontMatter = parseFrontMatter(markdownContent)

	// 提取元数据，优先使用表单中的标题，其次是Front Matter中的标题
	if title == "" && frontMatter.Title != "" {
		title = frontMatter.Title
	}

	if title == "" {
		e.Error(http.StatusBadRequest, nil, "请提供文章标题")
		return
	}

	// 将Markdown内容转换为HTML
	htmlContent := convertMarkdownToHTML(markdownContent)

	// 保存到数据库
	userID := user.GetUserId(c)
	articleID, err := e.saveArticleToDatabase(c, title, markdownContent, htmlContent, frontMatter, userID)
	if err != nil {
		e.Error(http.StatusInternalServerError, err, "保存文章失败")
		return
	}

	// 返回文章ID和标题
	e.OK(ArticleResponse{
		ID:      articleID,
		Title:   title,
		Message: "文章发布成功",
	}, "发布成功")
}

// parseFrontMatter 从Markdown内容中解析Front Matter
func parseFrontMatter(content string) FrontMatter {
	var frontMatter FrontMatter

	// 匹配Front Matter部分
	frontMatterRegex := regexp.MustCompile(`^---\s*([\s\S]*?)\s*---`)
	match := frontMatterRegex.FindStringSubmatch(content)

	if len(match) > 1 {
		frontMatterContent := match[1]
		lines := strings.Split(frontMatterContent, "\n")

		// 简单解析YAML格式的Front Matter
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 移除引号
			value = strings.Trim(value, `"'`)

			switch key {
			case "title":
				frontMatter.Title = value
			case "description":
				frontMatter.Description = value
			case "author":
				frontMatter.Author = value
			case "date":
				frontMatter.Date = value
			case "tags":
				// 简单解析标签数组
				value = strings.Trim(value, "[]")
				tags := strings.Split(value, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
					tags[i] = strings.Trim(tags[i], `"'`)
				}
				frontMatter.Tags = tags
			}
		}
	}

	return frontMatter
}

// convertMarkdownToHTML 将Markdown内容转换为HTML
func convertMarkdownToHTML(content string) string {
	// 使用gomarkdown库进行Markdown解析

	// 移除Front Matter部分
	frontMatterRegex := regexp.MustCompile(`^---\s*[\s\S]*?\s*---\s*`)
	content = frontMatterRegex.ReplaceAllString(content, "")

	// 创建Markdown解析器和HTML渲染器
	// 这里使用两种常见的库，根据项目需要选择一种

	// 方法1: 使用gomarkdown库
	mdParser := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	htmlRenderer := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})
	htmlBytes := markdown.ToHTML([]byte(content), mdParser, htmlRenderer)
	return string(htmlBytes)

	// 方法2: 使用goldmark库（如果项目中使用这个库）
	/*
		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
			),
		)
		var buf bytes.Buffer
		if err := md.Convert([]byte(content), &buf); err != nil {
			log.Printf("Markdown转换失败: %v", err)
			return content // 转换失败时返回原始内容
		}
		return buf.String()
	*/
}

// saveArticleToDatabase 保存文章到数据库
func (e *ArticleHandler) saveArticleToDatabase(c *gin.Context, title, markdownContent, htmlContent string, frontMatter FrontMatter, userID int) (int, error) {
	// 初始化上下文和ORM
	err := e.MakeContext(c).MakeOrm().Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, "初始化失败")
		return 0, err
	}

	// 获取ORM实例
	db, err := e.GetOrm()
	if err != nil {
		e.Logger.Error("获取数据库连接失败", "error", err)
		return 0, err
	}

	// 获取当前用户ID
	userId := int64(userID)

	// 直接使用ORM创建文章记录，避免服务层初始化问题
	article := &models.ArticleContent{
		Title:       title,
		Content:     markdownContent,
		HtmlContent: htmlContent,
		Status:      1, // 正常状态
	}

	// 设置创建者和更新者
	article.SetCreateBy(int(userId))
	article.SetUpdateBy(int(userId))

	// 保存到数据库，确保使用正确的表名
	err = db.Table("article_contents").Create(article).Error
	if err != nil {
		e.Logger.Error("保存文章到数据库失败", "error", err, "title", title)
		return 0, err
	}

	e.Logger.Info("文章保存到数据库成功", "title", title, "articleId", article.Id, "userId", userId)
	return int(article.Id), nil
}

// Preview 预览Markdown内容
// @Summary 预览Markdown内容
// @Description 将Markdown内容转换为HTML并返回
// @Tags 文章管理
// @Accept application/json,application/x-www-form-urlencoded
// @Produce json
// @Param content query string false "Markdown内容"
// @Success 200 {object} map[string]string "{"html": "<h1>HTML内容</h1>"}"
// @Router /api/v1/article/preview [get]
func (e *ArticleHandler) Preview(c *gin.Context) {
	// 获取Markdown内容
	content := c.Query("content")
	if content == "" {
		// 尝试从表单中获取
		content = c.PostForm("content")
	}

	if content == "" {
		e.Error(http.StatusBadRequest, nil, "请提供Markdown内容")
		return
	}

	// 转换为HTML
	htmlContent := convertMarkdownToHTML(content)

	e.OK(gin.H{"html": htmlContent}, "预览成功")
}

// truncateString 截断字符串，用于日志记录
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
