package resolved

type StatefulPlugin struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}
