package authlib

type User struct {
	UserID          string
	EmailAddress    string
	Name            string
	Password        string
	ConfirmPassword string
	Active          bool
	//OtherData       map[string]string
}
