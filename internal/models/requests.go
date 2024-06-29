package models

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetRequest struct {
	Key string `json:"key"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}
