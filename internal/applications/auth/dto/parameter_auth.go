package dto

type ParameterRegister struct {
	Email    string `json:"email"`
    Password string `json:"password"`
    FullName string `json:"full_name"`
    Phone    string `json:"phone"` // E.164 format: +628123456789
}
