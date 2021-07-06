package common

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UID string `json:"uid"`
}

type PingRequest struct {
	Text string `query:"text" json:"text"`
}

type PingResponse struct {
	Text string `json:"text"`
}
