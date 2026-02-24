package modelv2

type Resource struct {
	Name   string         `json:"name"`
	Role   string         `json:"role"`
	Config map[string]any `json:"config"`
}
