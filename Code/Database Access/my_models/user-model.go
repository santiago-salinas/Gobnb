package my_models

type User struct {
	ID    string `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	Roles []string `json:"roles" db:"roles"`
}



