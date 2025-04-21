package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MTranServerProvider struct {
	endpoints []string
}

func NewMTranServerProvider(endpoints []string) *MTranServerProvider {
	return &MTranServerProvider{
		endpoints: endpoints,
	}
}

func (p *MTranServerProvider) Translate(req TranslationRequest) (*TranslationResponse, error) {
	// 构建请求体
	requestBody := map[string]string{
		"from": req.From,
		"to":   req.To,
		"text": req.Text,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 尝试每个endpoint直到成功
	var lastErr error
	for _, endpoint := range p.endpoints {
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

		var result TranslationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			lastErr = err
			continue
		}

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
