package modelv2

type DataRetentionStoragePreferences struct {
	ProcErrors   bool `json:"procErrors"`
	GatewayDumps bool `json:"gatewayDumps"`
}

type DataRetention struct {
	DisableReportingPII bool                            `json:"disableReportingPii"`
	StoragePreferences  DataRetentionStoragePreferences `json:"storagePreferences"`
}

type Bucket struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type Settings struct {
	DataRetention   DataRetention   `json:"dataRetention"`
	UseSelfStorage  bool            `json:"useSelfStorage"`
	StorageBucket   Bucket          `json:"storageBucket"`
	RetentionPeriod RetentionPeriod `json:"retentionPeriod"`
}

type RetentionPeriod string

const (
	RetentionPeriodDefault RetentionPeriod = "default"
	RetentionPeriodFull    RetentionPeriod = "full"
)
