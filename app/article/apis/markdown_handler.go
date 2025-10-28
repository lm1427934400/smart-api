package apis

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownHandler 处理Markdown文件上传和内容处理
type MarkdownHandler struct {
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

// UploadMarkdownFile 处理Markdown文件上传
// @Summary 上传Markdown文件并发布文章
// @Description 上传Markdown文件并发布技术博客文章
// @Tags 文章管理
// @Accept multipart/form-data
// @Produce json
// @Param title formData string false "文章标题"
// @Param content formData string false "Markdown内容"
// @Param file formData file false ".md文件"
// @Success 200 {object} ArticleResponse "{\"code\": 0, \"data\": {\"id\": 1, \"title\": \"文章标题\"}}"
// @Failure 400 {object} response.Response "{\"code\": 400, \"message\": \"错误信息\"}"
// @Router /api/v1/sys-content [post]
// @Security Bearer
func (e MarkdownHandler) UploadMarkdownFile(c *gin.Context) {
	// 1. 验证文件类型和大小（限制50MB）
	file, header, err := c.Request.FormFile("file")
	var markdownContent string
	var frontMatter FrontMatter

	// 获取表单中的标题
	title := c.PostForm("title")

	// 如果提供了文件，则处理文件
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
			e.Error(http.StatusInternalServerError, err, "读取文件失败")
			return
		}

		markdownContent = string(fileBytes)

		// 2. 解析Front Matter
		frontMatter = parseFrontMatter(markdownContent)
	} else if err == http.ErrMissingFile {
		// 如果没有提供文件，尝试从表单获取内容
		markdownContent = c.PostForm("content")
		if markdownContent == "" {
			e.Error(http.StatusBadRequest, nil, "请提供文件或内容")
			return
		}

		// 解析Front Matter（如果存在）
		frontMatter = parseFrontMatter(markdownContent)
	} else {
		e.Error(http.StatusInternalServerError, err, "获取文件失败")
		return
	}

	// 3. 提取元数据，优先使用表单中的标题，其次是Front Matter中的标题
	if title == "" && frontMatter.Title != "" {
		title = frontMatter.Title
	}

	if title == "" {
		e.Error(http.StatusBadRequest, nil, "请提供文章标题")
		return
	}

	// 4. 将Markdown内容转换为HTML
	htmlContent := convertMarkdownToHTML(markdownContent)

	// 5. 保存到数据库
	userID := user.GetUserId(c)
	articleID, err := e.saveArticleToDatabase(title, markdownContent, htmlContent, frontMatter, userID)
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
func (e MarkdownHandler) saveArticleToDatabase(title, markdownContent, htmlContent string, frontMatter FrontMatter, userID int) (int, error) {
	// 使用与现有系统一致的方式创建文章记录
	// 这里我们使用简化的方式，假设系统有适当的模型和数据库连接方式
	
	// 创建一个临时的文章ID（实际应用中应该由数据库生成）
	// 在实际项目中，应该使用系统提供的数据库操作方法
	articleID := 1 // 模拟的文章ID
	
	// 这里应该是实际的数据库操作
	// 由于我们没有看到完整的数据库操作方式，暂时返回模拟的ID
	// 在实际实现中，应该使用与项目其他部分一致的数据库操作模式
	
	return articleID, nil
}

