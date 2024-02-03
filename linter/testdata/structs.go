package testdata

type TerraformResourceUnoptimized struct {
	Cloud               string // 16 bytes
	Name                string // 16 bytes
	HaveDSL             bool   //  1 byte
	PluginVersion       string // 16 bytes
	IsVersionControlled bool   //  1 byte
	TerraformVersion    string // 16 bytes
	ModuleVersionMajor  int32  //  4 bytes
}

type TerraformResourceOptimized struct {
	Cloud               string // 16 bytes
	Name                string // 16 bytes
	PluginVersion       string // 16 bytes
	TerraformVersion    string // 16 bytes
	ModuleVersionMajor  int32  //  4 bytes
	HaveDSL             bool   //  1 byte
	IsVersionControlled bool   //  1 byte
}
