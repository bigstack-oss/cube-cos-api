package auth

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (u User) IsNameEmpty() bool {
	return u.Name == ""
}

func (u User) IsPasswordEmpty() bool {
	return u.Password == ""
}
