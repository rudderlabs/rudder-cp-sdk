package model

import "time"

// WorkspaceConfigs represents workspace configurations of one or more workspaces, as well as definitions shared by all of them.
type WorkspaceConfigs struct {
	// Version is a version number of the current workspace config format, used for backward compatibility.
	// The current version is 2.
	Version int `json:"version"`
	// Workspaces is an map of workspace configurations. The key is a workspace ID.
	Workspaces map[string]*WorkspaceConfig `json:"workspaces"`
	// SourceDefinitions is a map of source definitions. The key is a source definition name.
	SourceDefinitions map[string]*SourceDefinition `json:"sourceDefinitions"`
	// DestinationDefinitions is a map of destination definitions. The key is a destination definition name.
	DestinationDefinitions map[string]*DestinationDefinition `json:"destinationDefinitions"`
}

type WorkspaceConfig struct {
	Settings                         *Settings                          `json:"settings"`
	Sources                          map[string]*Source                 `json:"sources"`
	Destinations                     map[string]*Destination            `json:"destinations"`
	Connections                      []*Connection                      `json:"connections"`
	Libraries                        []*Library                         `json:"libraries"`
	WHTProjects                      map[string]*WHTProject             `json:"whtProjects"`
	Accounts                         map[string]*Account                `json:"accounts"`
	Transformations                  map[string]*Transformation         `json:"transformations"`
	TrackingPlans                    map[string]*TrackingPlan           `json:"trackingPlans"`
	Resources                        map[string]*Resource               `json:"resources"`
	AudienceSources                  []*AudienceSource                  `json:"audienceSources"`
	SQLModelVersions                 map[string]*SQLModelVersion        `json:"sqlModelVersions"`
	SQLModelVersionSourceConnections []*SQLModelVersionSourceConnection `json:"sqlModelVersionSourceConnections"`
	UpdatedAt                        time.Time                          `json:"updatedAt"`
}

// Settings

type Settings struct {
	DataRetention struct {
		DisableReportingPII bool `json:"disableReportingPii"`
		StoragePreferences  struct {
			ProcErrors   bool `json:"procErrors"`
			GatewayDumps bool `json:"gatewayDumps"`
		} `json:"storagePreferences"`
	} `json:"dataRetention"`
	UseSelfStorage bool `json:"useSelfStorage"`
	StorageBucket  struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	} `json:"storageBucket"`
	RetentionPeriod RetentionPeriod `json:"retentionPeriod"`
}

type RetentionPeriod string

const (
	RetentionPeriodDefault RetentionPeriod = "default"
	RetentionPeriodFull    RetentionPeriod = "full"
)

// Sources

type Source struct {
	Name           string                 `json:"name"`
	WriteKey       string                 `json:"writeKey"`
	Enabled        bool                   `json:"enabled"`
	Deleted        bool                   `json:"deleted"`
	Config         map[string]interface{} `json:"config"`
	Transient      bool                   `json:"transient"`
	DefinitionName string                 `json:"sourceDefinitionName"`
	SecretVersion      int                         `json:"secretVersion"` // TODO: incompatible with the current type in dev
	AccountID          string                      `json:"accountId"`
	TrackingPlanConfig *DGSourceTrackingPlanConfig `json:"dgSourceTrackingPlanConfig"`
	// TODO: consider adding the following fields
	// CreatedBy string `json:"createdBy"`
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
	// LiveEventsConfig ...
}

type SourceDefinition struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Category    string                 `json:"category"`
	Options     map[string]interface{} `json:"options"`
	Config      map[string]interface{} `json:"config"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

type DGSourceTrackingPlanConfig struct {
	TrackingPlanID string `json:"trackingPlanId"`
	Version        int    `json:"version"`
	Config         struct {
		Global   *DGSourceTrackingPlanOptions `json:"global"`
		Track    *DGSourceTrackingPlanOptions `json:"track"`
		Identify *DGSourceTrackingPlanOptions `json:"identify"`
		Group    *DGSourceTrackingPlanOptions `json:"group"`
	}
	Deleted bool `json:"deleted"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

type DGSourceTrackingPlanOptions struct {
	AllowUnplannedEvents interface{}            `json:"allowUnplannedEvents"`
	SendViolatedEventsTo string                 `json:"sendViolatedEventsTo"`
	AJVOptions           map[string]interface{} `json:"ajvOptions"`
}

// Destinations

type Destination struct {
	Name              string                 `json:"name"`
	Enabled           bool                   `json:"enabled"`
	Config            map[string]interface{} `json:"config"`
	DefinitionName    string                 `json:"destinationDefinitionName"`
	Deleted           bool                   `json:"deleted"`
	TransformationIDs []string               `json:"transformationIds"`
	RevisionID        string                 `json:"revisionId"`
	SecretVersion     int                    `json:"secretVersion"` // TODO: incompatible with the current type in dev
	Regions []*DestinationRegion `json:"regions"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
	// LiveEventsConfig ...
}

type Region string

const (
	RegionUS Region = "US"
	RegionEU Region = "EU"
)

type DestinationRegion struct {
	Region     Region                 `json:"region"`
	Config     map[string]interface{} `json:"config"`
	RevisionID string                 `json:"revisionId"`
	SecretVersion int            `json:"secretVersion"` // TODO: incompatible with the current type in dev
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

type DestinationDefinition struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Category    string                 `json:"category"`
	Options     map[string]interface{} `json:"options"`
	Config      map[string]interface{} `json:"config"`
	// TODO: consider adding the following fields
	// ResponseRules map[string]interface{} `json:"responseRules"`
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

// Connections

type Connection struct {
	SourceID         string `json:"sourceId"`
	DestinationID    string `json:"destinationId"`
	Enabled          bool   `json:"enabled"`
	ProcessorEnabled bool   `json:"processorEnabled"`
}

// Libraries

type Library struct {
	VersionID string `json:"versionId"`
}

// Transformations

type Transformation struct {
	VersionID string `json:"versionId"`
	// TODO: consider adding the following fields
	// LiveEventsConfig ...
}

// Accounts

type Account struct {
	Name     string                 `json:"name"`
	Role     string                 `json:"role"`
	Options  map[string]interface{} `json:"options"`
	Secret   map[string]interface{} `json:"secret"`
	Metadata map[string]interface{} `json:"metadata"`
	UserID   string                 `json:"userId"`
	SecretVersion int                    `json:"secretVersion"` // TODO: incompatible with the current type in dev
	Category AccountCategory `json:"rudderCategory"`
}

type AccountCategory string

const (
	AccountCategorySource        AccountCategory = "source"
	AccountCategoryDestination   AccountCategory = "destination"
	AccountCategoryDataRetention AccountCategory = "dataRetention"
	AccountCategoryWHT           AccountCategory = "wht"
)

// WHTProjects

type WHTProject struct {
	Name       string   `json:"name"`
	Descrition string   `json:"description"`
	Config     struct { // TODO: these should not be arbitrary maps
		Resources       []string               `json:"resources"`
		Schedule        map[string]interface{} `json:"schedule"`
		RetentionPolicy map[string]interface{} `json:"retentionPolicy"`
	} `json:"config"`
	SSHKeyID           string `json:"sshKeyId"`
	Enabled            bool   `json:"enabled"`
	Deleted            bool   `json:"deleted"`
	GitRepoAccountID   string `json:"gitRepoAccountId"`
	WarehouseAccountID string `json:"warehouseAccountId"`
}

// tracking plans

type TrackingPlan struct {
	Version int `json:"version"`
}

// resources

type Resource struct {
	Name   string                 `json:"name"`
	Role   string                 `json:"role"`
	Config map[string]interface{} `json:"config"`
}

// audience source

type AudienceSource struct {
	ID        string `json:"id"`
	SourceID  string `json:"sourceId"`
	AccountID string `json:"accountId"`
	Deleted   bool   `json:"deleted"`
}

// sql model version

type SQLModelVersion struct {
	SQLModelID string `json:"sqlModelId"`
	AccountID  string `json:"accountId"`
}

type SQLModelVersionSourceConnection struct {
	SourceID          string `json:"sourceId"`
	SQLModelVersionID string `json:"sqlModelVersionId"`
}
