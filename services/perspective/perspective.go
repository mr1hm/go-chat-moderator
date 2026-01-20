package perspective

type PerspectiveRequest struct {
	Comment struct {
		Text string `json:"text"`
	} `json:"comment"`
	RequestedAttributes map[string]interface{} `json:"requestedAttributes"`
}

type PerspectiveResponse struct {
	AttributesScores map[string]struct {
		SummaryScore struct {
			Value float64 `json:"value"`
		} `json:"summaryScore"`
	} `json:"attributeScores"`
}
