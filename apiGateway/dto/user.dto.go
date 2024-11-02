package dto

type RegisterDto struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type LoginDto struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
}
