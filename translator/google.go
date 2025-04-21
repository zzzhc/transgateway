package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GoogleProvider struct {
	proxy string
}

func NewGoogleProvider(proxy string) *GoogleProvider {
	return &GoogleProvider{
		proxy: proxy,
	}
}

func (p *GoogleProvider) Translate(req TranslationRequest) (*TranslationResponse, error) {
	params := url.Values{}
	params.Set("client", "gtx")
	params.Set("sl", req.From)
	params.Set("tl", req.To)
	params.Set("dt", "t")
	params.Set("q", req.Text)

	client := &http.Client{}
	if p.proxy != "" {
		proxyURL, err := url.Parse(p.proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Get(fmt.Sprintf("https://translate.googleapis.com/translate_a/single?%s", params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Printf("Google翻译API响应: %s, code=%d\n", string(body), resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// [[["你好","hello",null,null,10]],null,"en",null,null,null,null,[]]
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	var translatedText strings.Builder
	if len(result) > 0 {
		if translations, ok := result[0].([]interface{}); ok {
			for _, translation := range translations {
				if sentenceArray, ok := translation.([]interface{}); ok && len(sentenceArray) > 1 {
					if text, ok := sentenceArray[0].(string); ok {
						translatedText.WriteString(text)
					}
				}
			}
		}
	}

	detectedLang := "auto"
	if len(result) > 2 {
		if lang, ok := result[2].(string); ok {
			detectedLang = lang
		}
	}

	return &TranslationResponse{
		DetectedSourceLang: detectedLang,
		Result:             translatedText.String(),
	}, nil
}

func (p *GoogleProvider) BatchTranslate(req BatchTranslationRequest) (*BatchTranslationResponse, error) {
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
