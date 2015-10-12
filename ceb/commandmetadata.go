package ceb

type CommandParameter struct {
	Key         string `json:"key"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
}

type CommandMap struct {
	Name       string             `json:"name"`
	Code       string             `json:"code"`
	Parameters []CommandParameter `json:"parameters"`
}

type StatMetadata struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	XAxis string `json:"xAxis"`
	YAxis string `json:"yAxis"`
	MaxX  int    `json:"maxX"`
}

type ConfigMetadata struct {
	Name       string                 `json:"name"`
	Code       string                 `json:"code"`
	Parameters map[string]interface{} `json:"parameters"`
}
