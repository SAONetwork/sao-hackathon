package server

type TextRazorResponse struct {
	Response struct {
		CoarseTopics []struct {
			ID       int    `json:"id"`
			Label    string `json:"label"`
			WikiLink string `json:"wikiLink"`
			Score    int    `json:"score"`
		} `json:"coarseTopics"`
		Language           string `json:"language"`
		LanguageIsReliable bool   `json:"languageIsReliable"`
	} `json:"response"`
	Time float64 `json:"time"`
	Ok   bool    `json:"ok"`
}