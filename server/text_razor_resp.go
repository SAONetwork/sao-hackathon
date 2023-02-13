package server

type TextRazorResponse struct {
	Response struct {
		CoarseTopics []struct {
			ID       int    `json:"id"`
			Label    string `json:"label"`
		} `json:"coarseTopics"`
		Language           string `json:"language"`
		LanguageIsReliable bool   `json:"languageIsReliable"`
		Topics             []struct {
			ID         int     `json:"id"`
			Label      string  `json:"label"`
		} `json:"topics"`
	} `json:"response"`
	Time float64 `json:"time"`
	Ok   bool    `json:"ok"`
}