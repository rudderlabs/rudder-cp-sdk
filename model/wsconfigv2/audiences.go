package wsconfigv2

type AudienceSource struct {
	ID        string `json:"id"`
	SourceID  string `json:"sourceId"`
	AccountID string `json:"accountId"`
	Deleted   bool   `json:"deleted"`
}
