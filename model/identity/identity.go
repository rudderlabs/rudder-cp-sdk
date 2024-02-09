package identity

type AdminCredentials struct {
	AdminUsername string
	AdminPassword string
}

type Workspace struct {
	WorkspaceToken string
}

type Namespace struct {
	Namespace string
	Secret    string
}
