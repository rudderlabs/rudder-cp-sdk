package notifications

// WorkspaceConfigNotification is a notification that the workspace configs have been updated.
// This is intentionally empty for now because it is expected that consumers will always
// get the latest configs from ControlPlane SDK using GetWorkspaceConfigs.
// In the future, this may be used to notify consumers of specific changes to the configs.
type WorkspaceConfigNotification struct{}
