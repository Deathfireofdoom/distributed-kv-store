package models

type PutResponse struct {
	Success bool `json:"success"`
}

type GetResponse struct {
	Value string `json:"value"`
	Found bool   `json:"found"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
}
