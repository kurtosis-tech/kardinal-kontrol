package resolved

type StatefulPlugin struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	ServiceName string            `json:"servicename"`
	Args        map[string]string `json:"args"`
}
