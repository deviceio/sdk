package hmapi

type Content struct {
	Type  MediaType   `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type ContentRequest interface{}

type contentRequest struct {
}
