package modelv2

type Connection struct {
	SourceID         string `json:"sourceId"`
	DestinationID    string `json:"destinationId"`
	Enabled          bool   `json:"enabled"`
	ProcessorEnabled bool   `json:"processorEnabled"`
}
