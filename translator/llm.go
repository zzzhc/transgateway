package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// LLMConfig 定义了LLM提供者的配置
type LLMConfig struct {
	BaseURL      string
	APIKey       string
	Model        string
	SystemPrompt string
	UserPrompt   string
}

// LLMProvider 实现了基于LLM的翻译提供者
type LLMProvider struct {
	config LLMConfig
	client openai.Client
}

// NewLLMProvider 创建一个新的LLM翻译提供者
func NewLLMProvider(config LLMConfig) (*LLMProvider, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.BaseURL),
	)

	return &LLMProvider{
		config: config,
		client: client,
	}, nil
}

// Translate 实现单个文本的翻译
func (p *LLMProvider) Translate(req TranslationRequest) (*TranslationResponse, error) {
	// 构建用户提示
	userPrompt := p.config.UserPrompt
	if userPrompt == "" {
		userPrompt = "Translate to {{to}}. Keep untranslatable parts (like proper nouns, code) as is. *Output ONLY the translation*:\n{{text}}"
	}
	userPrompt = strings.ReplaceAll(userPrompt, "{{to}}", req.To)
	userPrompt = strings.ReplaceAll(userPrompt, "{{text}}", req.Text)

	// 构建系统提示
	systemPrompt := p.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a professional translation engine."
	}

	// 调用LLM API
	chatCompletion, err := p.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model: p.config.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API调用失败: %v", err)
	}

	return &TranslationResponse{
		Result: chatCompletion.Choices[0].Message.Content,
	}, nil
}

// BatchTranslate 实现批量文本的翻译
func (p *LLMProvider) BatchTranslate(req BatchTranslationRequest) (*BatchTranslationResponse, error) {
	results := make([]string, len(req.Texts))
	var detectedLang string

	for i, text := range req.Texts {
		resp, err := p.Translate(TranslationRequest{
			From: req.From,
			To:   req.To,
			Text: text,
		})
		if err != nil {
			return nil, err
		}
		results[i] = resp.Result
		if i == 0 {
			detectedLang = resp.DetectedSourceLang
		}
	}

	return &BatchTranslationResponse{
		DetectedSourceLang: detectedLang,
		Results:            results,
	}, nil
}
