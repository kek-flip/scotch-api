package model

type User struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
	Name              string `json:"name"`
	Age               int    `json:"age"`
	Gender            string `json:"gender"`
	PhoneNumber       string `json:"phone_number"`
	About             string `json:"about"`
}
