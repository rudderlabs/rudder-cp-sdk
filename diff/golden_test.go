package diff

import "time"

var goldenWorkspace1 = &WorkspaceConfig{
	Sources: map[string]*Source{
		"source1": {
			Name:           "Dev Integration Test 1",
			WriteKey:       "writeKey1",
			Enabled:        true,
			DefinitionName: "webhook",
			Config: []byte(`{
						"k1": "v1",
						"k2": 123,
						"k3": true,
						"k4": 123.123
					}`),
		},
	},
	Destinations: map[string]*Destination{
		"destination1": {Enabled: true},
	},
	Connections: map[string]*Connection{
		"connectionID1": {SourceID: "source1", DestinationID: "destination1", ProcessorEnabled: true},
	},
	UpdatedAt: time.Date(2021, 9, 1, 1, 2, 3, 0, time.UTC),
}

var goldenWorkspace2 = &WorkspaceConfig{
	Sources: map[string]*Source{
		"source2": {
			Name:           "Dev Integration Test 2",
			WriteKey:       "writeKey2",
			Enabled:        true,
			DefinitionName: "webhook",
			Config: []byte(`{
						"k5": "v1",
						"k6": 123,
						"k7": true,
						"k8": 123.123
					}`),
		},
	},
	Destinations: map[string]*Destination{
		"destination2": {Enabled: true},
	},
	Connections: map[string]*Connection{
		"connectionID2": {SourceID: "source2", DestinationID: "destination2", ProcessorEnabled: true},
	},
	UpdatedAt: time.Date(2021, 9, 1, 6, 6, 6, 0, time.UTC),
}

var goldenWorkspace3 = &WorkspaceConfig{
	Sources: map[string]*Source{
		"source3": {
			Name:           "Dev Integration Test 3",
			WriteKey:       "writeKey3",
			Enabled:        true,
			DefinitionName: "webhook",
			Config: []byte(`{
						"k9": "v1",
						"k0": 123
					}`),
		},
	},
	Destinations: map[string]*Destination{
		"destination3": {Enabled: true},
	},
	Connections: map[string]*Connection{
		"connectionID3": {SourceID: "source3", DestinationID: "destination3", ProcessorEnabled: true},
	},
	UpdatedAt: time.Date(2021, 9, 1, 6, 6, 7, 0, time.UTC),
}

var goldenUpdatedWorkspace3 = &WorkspaceConfig{
	Sources: map[string]*Source{
		"source3": {
			Name:           "Dev Integration Test 3 - DISABLED",
			WriteKey:       "writeKey3",
			Enabled:        false,
			DefinitionName: "webhook",
			Config: []byte(`{
						"k9": "v1",
						"k0": 123
					}`),
		},
	},
	Destinations: map[string]*Destination{
		"destination3": {Enabled: true},
	},
	Connections: map[string]*Connection{
		"connectionID3": {SourceID: "source3", DestinationID: "destination3", ProcessorEnabled: true},
	},
	UpdatedAt: time.Date(2021, 9, 1, 6, 6, 8, 0, time.UTC),
}
