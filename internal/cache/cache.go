package cache

import (
	"github.com/rudderlabs/rudder-cp-sdk/subscriber"
	"sync"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/notifications"
)

type WorkspaceConfigCache struct {
	configs        *modelv2.WorkspaceConfigs
	updateLock     sync.Mutex
	subscribers    []*subscriber.Subscriber
	subscriberLock sync.Mutex
}

type WorkspaceConfigNotification struct{}

// Get returns a copy of the current workspace configs.
// This is in order to avoid race conditions when consumers read the configs while they are being updated.
func (c *WorkspaceConfigCache) Get() *modelv2.WorkspaceConfigs {
	c.updateLock.Lock()
	defer c.updateLock.Unlock()

	if c.configs == nil {
		return nil
	}

	return copyConfigs(c.configs)
}

// Set updates the current workspace configs by merging input with current cache contents.
// Workspace configs are merged by id, so if a workspace config is updated, it will replace the previous one.
// If a workspace config is not included in input, it will be removed from the cache.
// If a workspace config is nil in input, it will not be updated.
// Source and destination definitions are merged without removing any missing definitions.
// It notifies all subscribers of the update.
func (c *WorkspaceConfigCache) Set(configs *modelv2.WorkspaceConfigs) {
	c.updateLock.Lock()
	c.merge(configs)
	c.updateLock.Unlock()

	c.subscriberLock.Lock()
	defer c.subscriberLock.Unlock()

	for _, s := range c.subscribers {
		s.Notify(notifications.WorkspaceConfigNotification{})
	}
}

func (c *WorkspaceConfigCache) merge(configs *modelv2.WorkspaceConfigs) {
	if c.configs == nil {
		c.configs = modelv2.Empty()
	}

	// merge source and destination definitions
	for id, config := range configs.SourceDefinitions {
		c.configs.SourceDefinitions[id] = config
	}

	for id, config := range configs.DestinationDefinitions {
		c.configs.DestinationDefinitions[id] = config
	}

	// remove deleted workspace configs (missing ids)
	currentIds := make([]string, 0, len(c.configs.Workspaces))
	for id := range c.configs.Workspaces {
		currentIds = append(currentIds, id)
	}

	for _, id := range currentIds {
		if _, ok := configs.Workspaces[id]; !ok {
			delete(c.configs.Workspaces, id)
		}
	}

	// merge workspace configs with updates (not nil values)
	for id, config := range configs.Workspaces {
		if config != nil {
			c.configs.Workspaces[id] = config
		}
	}
}

// Subscribe returns a subscriber that will be notified of any updates to the workspace configs.
// Subscribers are notified in the order they are subscribed.
// They can monitor for updates by reading from the notifications channel, provided by the Notifications function.
// It is expected to handle any notifications in a timely manner, otherwise it will block the cache from updating.
func (c *WorkspaceConfigCache) Subscribe() *subscriber.Subscriber {
	c.subscriberLock.Lock()
	defer c.subscriberLock.Unlock()

	s := subscriber.New()
	c.subscribers = append(c.subscribers, s)

	return s
}

func copyConfigs(c *modelv2.WorkspaceConfigs) *modelv2.WorkspaceConfigs {
	wc := &modelv2.WorkspaceConfigs{
		SourceDefinitions:      make(map[string]*modelv2.SourceDefinition),
		DestinationDefinitions: make(map[string]*modelv2.DestinationDefinition),
		Workspaces:             make(map[string]*modelv2.WorkspaceConfig),
	}

	for k, v := range c.SourceDefinitions {
		wc.SourceDefinitions[k] = v
	}

	for k, v := range c.DestinationDefinitions {
		wc.DestinationDefinitions[k] = v
	}

	for k, v := range c.Workspaces {
		wc.Workspaces[k] = v
	}

	return wc
}
