package modelv2

type SQLModelVersion struct {
	SQLModelID string `json:"sqlModelId"`
	AccountID  string `json:"accountId"`
}

type SQLModelVersionSourceConnection struct {
	SourceID          string `json:"sourceId"`
	SQLModelVersionID string `json:"sqlModelVersionId"`
}
