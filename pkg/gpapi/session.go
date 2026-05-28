package gpapi

type Session struct {
	AndroidId float64 `json:"androidId"`
	Email     string  `json:"email"`
	AasToken  string  `json:"aasToken"`
	AdmToken  string  `json:"admToken"`
}
