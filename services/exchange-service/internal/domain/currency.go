package domain

type Currency struct {
	Code      string `json:"code"`
	Precision int    `json:"precision"`
}

type Rate struct {
	Base       string `json:"base"`
	Quote      string `json:"quote"`
	ObservedOn string `json:"date"`
	Value      string `json:"rate"`
	Source     string `json:"source"`
}
