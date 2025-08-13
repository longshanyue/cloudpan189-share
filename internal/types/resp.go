package types

type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
