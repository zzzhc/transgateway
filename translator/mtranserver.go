package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/pemistahl/lingua-go"
)

type MTranServerProvider struct {
	endpoints            []string
	detector             lingua.LanguageDetector
	currentEndpointIndex int64
}

func NewMTranServerProvider(endpoints []string) *MTranServerProvider {
	// 创建语言检测器，支持所有语言
	detector := lingua.NewLanguageDetectorBuilder().
		FromAllLanguages().
		Build()

	return &MTranServerProvider{
		endpoints:            endpoints,
		detector:             detector,
		currentEndpointIndex: 0,
	}
}

func (p *MTranServerProvider) getNextEndpoint() string {
	if len(p.endpoints) == 0 {
		return ""
	}
	// 使用原子操作获取当前索引
	index := atomic.AddInt64(&p.currentEndpointIndex, 1) - 1
	// 确保索引在有效范围内
	index = index % int64(len(p.endpoints))
	if index < 0 {
		index = 0
	}
	return p.endpoints[index]
}

func (p *MTranServerProvider) Translate(req TranslationRequest) (*TranslationResponse, error) {
	if len(p.endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}
	// 如果 from 是 auto，则检测语言
	fromLang := req.From
	detectedLang := ""
	if fromLang == "auto" {
		if language, exists := p.detector.DetectLanguageOf(req.Text); exists {
			// 将 lingua 的语言代码转换为翻译服务使用的语言代码
			fromLang = strings.ToLower(language.IsoCode639_1().String())
		} else {
			fromLang = "en" // 如果无法检测，默认使用英语
		}
		detectedLang = fromLang
	}

	// 检查是否需要两步翻译（当目标语言是中文且源语言不是英文时）
	needTwoStepTranslation := !(req.To == "en" || fromLang == "en")

	var intermediateResult *TranslationResponse
	var err error

	if needTwoStepTranslation {
		// 第一步：先翻译到英文
		intermediateResult, err = p.translateWithEndpoints(detectedLang, fromLang, "en", req.Text)
		if err != nil {
			return nil, fmt.Errorf("failed to translate to English: %v", err)
		}
		// 第二步：从英文翻译到目标语言
		return p.translateWithEndpoints(detectedLang, "en", req.To, intermediateResult.Result)
	}

	// 直接翻译
	return p.translateWithEndpoints(detectedLang, fromLang, req.To, req.Text)
}

// translateWithEndpoints 是实际的翻译实现，处理与endpoint的通信
func (p *MTranServerProvider) translateWithEndpoints(detectedLang, from, to, text string) (*TranslationResponse, error) {
	requestBody := map[string]string{
		"from": from,
		"to":   to,
		"text": text,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	var lastErr error
	attemptedEndpoints := make(map[string]bool)

	for len(attemptedEndpoints) < len(p.endpoints) {
		endpoint := p.getNextEndpoint()
		if attemptedEndpoints[endpoint] {
			continue
		}
		attemptedEndpoints[endpoint] = true

		resp, err := http.Post(
			fmt.Sprintf("%s/translate", endpoint),
			"application/json; charset=utf-8",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP error: %d", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var result TranslationResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}
		result.DetectedSourceLang = detectedLang

		return &result, nil
	}

	return nil, fmt.Errorf("all endpoints failed, last error: %v", lastErr)
}

func (p *MTranServerProvider) BatchTranslate(req BatchTranslationRequest) (*BatchTranslationResponse, error) {
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
