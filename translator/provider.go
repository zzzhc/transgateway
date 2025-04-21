package translator

type TranslationRequest struct {
	From string
	To   string
	Text string
}

type TranslationResponse struct {
	DetectedSourceLang string `json:"detectedSourceLang,omitempty"`
	Result             string `json:"result"`
}

type BatchTranslationRequest struct {
	From  string   `json:"from"`
	To    string   `json:"to"`
	Texts []string `json:"texts"`
}

type BatchTranslationResponse struct {
	DetectedSourceLang string   `json:"detectedSourceLang,omitempty"`
	Results            []string `json:"results"`
}

type Provider interface {
	Translate(req TranslationRequest) (*TranslationResponse, error)
	BatchTranslate(req BatchTranslationRequest) (*BatchTranslationResponse, error)
}
