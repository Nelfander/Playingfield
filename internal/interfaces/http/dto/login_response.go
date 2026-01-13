package dto

type LoginResponse struct {
	Token  string       `json:"token"`
	UserId int64        `json:"userId"`
	User   UserResponse `json:"user"`
}
