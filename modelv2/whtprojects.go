package modelv2

type WHTProjectConfig struct { // TODO: these should not be arbitrary maps
	Resources       []string               `json:"resources"`
	Schedule        map[string]interface{} `json:"schedule"`
	RetentionPolicy map[string]interface{} `json:"retentionPolicy"`
}

type WHTProject struct {
	Name               string           `json:"name"`
	Descrition         string           `json:"description"`
	Config             WHTProjectConfig `json:"config"`
	SSHKeyID           string           `json:"sshKeyId"`
	Enabled            bool             `json:"enabled"`
	Deleted            bool             `json:"deleted"`
	GitRepoAccountID   string           `json:"gitRepoAccountId"`
	WarehouseAccountID string           `json:"warehouseAccountId"`
}
