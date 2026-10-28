package domain

type (
	AdminRepository interface {
		Register(username string, password string) error
		Login(username string) (Admin, error)
	}
	Admin struct {
		AdminId  string `json:"id,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
)
