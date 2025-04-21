package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"translator/config"
	"translator/translator"

	"github.com/gin-gonic/gin"
)

//go:embed static/playground.html
var staticFS embed.FS

var (
	configPath = flag.String("c", "config.yml", "配置文件路径")
	debugMode  = flag.Bool("d", false, "启用调试模式")
)

func main() {
	flag.Parse()

	if !*debugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化翻译提供者
	providers := make(map[string]translator.Provider)

	// 添加 OpenCC 提供者
	providers["opencc"] = translator.NewOpenCCProvider()

	// 初始化其他提供者
	for name, provider := range cfg.Providers {
		if !provider.Enable {
			continue
		}

		switch name {
		case "mtranserver":
			providers[name] = translator.NewMTranServerProvider(provider.Endpoints)
		case "google":
			providers[name] = translator.NewGoogleProvider(provider.Proxy)
		default:
			if provider.LLM {
				llmConfig := translator.LLMConfig{
					BaseURL:      provider.BaseUrl,
					APIKey:       provider.ApiKey,
					Model:        provider.Model,
					SystemPrompt: provider.SystemPrompt,
					UserPrompt:   provider.UserPrompt,
				}
				llmProvider, err := translator.NewLLMProvider(llmConfig)
				if err != nil {
					log.Printf("初始化LLM提供者 %s 失败: %v", name, err)
					continue
				}
				providers[name] = llmProvider
			}
		}
	}

	// 设置Gin路由
	r := gin.Default()

	// 静态文件服务
	playgroundHTML, err := staticFS.ReadFile("static/playground.html")
	if err != nil {
		log.Fatalf("读取内嵌文件失败: %v", err)
	}
	r.GET("/play", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html", playgroundHTML)
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 获取可用的翻译提供者列表
	r.GET("/providers", func(c *gin.Context) {
		providerList := make([]string, 0, len(providers))
		for name := range providers {
			providerList = append(providerList, name)
		}
		c.JSON(http.StatusOK, gin.H{
			"providers": providerList,
		})
	})

	// 单次翻译
	r.GET("/translate", func(c *gin.Context) {
		handleTranslate(c, providers)
	})

	r.POST("/translate", func(c *gin.Context) {
		handleTranslate(c, providers)
	})

	// 批量翻译
	r.POST("/batch_translate", func(c *gin.Context) {
		var req struct {
			Provider string   `json:"provider"`
			From     string   `json:"from"`
			To       string   `json:"to"`
			Texts    []string `json:"texts"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Provider == "" {
			req.Provider = "mtranserver"
		}

		provider, ok := providers[req.Provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("不支持的翻译服务: %s", req.Provider)})
			return
		}

		result, err := provider.BatchTranslate(translator.BatchTranslationRequest{
			From:  req.From,
			To:    req.To,
			Texts: req.Texts,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// 繁简转换
	r.GET("/cc", func(c *gin.Context) {
		handleCC(c, providers["opencc"])
	})

	r.POST("/cc", func(c *gin.Context) {
		handleCC(c, providers["opencc"])
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("服务器启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func handleTranslate(c *gin.Context, providers map[string]translator.Provider) {
	var req struct {
		Provider string `json:"provider" form:"provider"`
		From     string `json:"from" form:"from"`
		To       string `json:"to" form:"to"`
		Text     string `json:"text" form:"text"`
	}

	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		req.Provider = c.Query("provider")
		req.From = c.Query("from")
		req.To = c.Query("to")
		req.Text = c.Query("text")
	}

	if req.Provider == "" {
		req.Provider = "mtranserver"
	}

	provider, ok := providers[req.Provider]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("不支持的翻译服务: %s", req.Provider)})
		return
	}

	result, err := provider.Translate(translator.TranslationRequest{
		From: req.From,
		To:   req.To,
		Text: req.Text,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleCC 处理繁简转换请求
func handleCC(c *gin.Context, provider translator.Provider) {
	var req struct {
		Type string `json:"type" form:"type"`
		Text string `json:"text" form:"text"`
	}

	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		req.Type = c.Query("type")
		req.Text = c.Query("text")
	}

	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 参数不能为空"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text 参数不能为空"})
		return
	}

	// 解析转换类型
	var from, to string
	switch req.Type {
	case "s2t":
		from, to = "s", "t"
	case "t2s":
		from, to = "t", "s"
	case "s2tw":
		from, to = "s", "tw"
	case "tw2s":
		from, to = "tw", "s"
	case "s2hk":
		from, to = "s", "hk"
	case "hk2s":
		from, to = "hk", "s"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的转换类型"})
		return
	}

	result, err := provider.Translate(translator.TranslationRequest{
		From: from,
		To:   to,
		Text: req.Text,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
