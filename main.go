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
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化翻译提供者
	providers := make(map[string]translator.Provider)

	if cfg.Providers.Mtranserver.Enable {
		providers["mtranserver"] = translator.NewMTranServerProvider(cfg.Providers.Mtranserver.Endpoints)
	}

	if cfg.Providers.Google.Enable {
		providers["google"] = translator.NewGoogleProvider(cfg.Providers.Google.Proxy)
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
