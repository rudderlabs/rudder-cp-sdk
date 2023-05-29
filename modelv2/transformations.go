package modelv2

type Library struct {
	VersionID string `json:"versionId"`
}

type Transformation struct {
	VersionID string `json:"versionId"`
	// TODO: consider adding the following fields
	// LiveEventsConfig ...
}
