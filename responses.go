package ispend

type APIResponse struct {
	Status  int         `json:"status"`
	IsError bool        `json:"isError"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
