package translator

import (
	"fmt"
	"sync"

	"github.com/liuzl/gocc"
)

// OpenCCProvider 实现繁简转换
type OpenCCProvider struct {
	// 使用 sync.Map 来存储转换器缓存，key 为转换类型
	converters sync.Map
	// 用于保护创建转换器的互斥锁
	mu sync.Mutex
}

// NewOpenCCProvider 创建新的 OpenCC 提供者
func NewOpenCCProvider() *OpenCCProvider {
	return &OpenCCProvider{}
}

// getConverter 获取或创建转换器
func (p *OpenCCProvider) getConverter(conversionType string) (*gocc.OpenCC, error) {
	// 先从缓存中获取
	if converter, ok := p.converters.Load(conversionType); ok {
		return converter.(*gocc.OpenCC), nil
	}

	// 缓存未命中，需要创建新的转换器
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查，避免重复创建
	if converter, ok := p.converters.Load(conversionType); ok {
		return converter.(*gocc.OpenCC), nil
	}

	// 创建新的转换器
	converter, err := gocc.New(conversionType)
	if err != nil {
		return nil, fmt.Errorf("创建转换器失败: %v", err)
	}

	// 存入缓存
	p.converters.Store(conversionType, converter)
	return converter, nil
}

// Translate 实现繁简转换
func (p *OpenCCProvider) Translate(req TranslationRequest) (*TranslationResponse, error) {
	if req.From == "" || req.To == "" {
		return nil, fmt.Errorf("from 和 to 参数不能为空")
	}

	// 构建转换类型
	conversionType := req.From + "2" + req.To
	converter, err := p.getConverter(conversionType)
	if err != nil {
		return nil, err
	}

	result, err := converter.Convert(req.Text)
	if err != nil {
		return nil, fmt.Errorf("转换失败: %v", err)
	}

	return &TranslationResponse{
		Result: result,
	}, nil
}

// BatchTranslate 批量转换
func (p *OpenCCProvider) BatchTranslate(req BatchTranslationRequest) (*BatchTranslationResponse, error) {
	if req.From == "" || req.To == "" {
		return nil, fmt.Errorf("from 和 to 参数不能为空")
	}

	// 构建转换类型
	conversionType := req.From + "2" + req.To
	converter, err := p.getConverter(conversionType)
	if err != nil {
		return nil, err
	}

	results := make([]string, len(req.Texts))
	for i, text := range req.Texts {
		result, err := converter.Convert(text)
		if err != nil {
			return nil, fmt.Errorf("转换文本 '%s' 失败: %v", text, err)
		}
		results[i] = result
	}

	return &BatchTranslationResponse{
		DetectedSourceLang: req.From,
		Results:            results,
	}, nil
}
