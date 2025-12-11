package entity

type UserEntity struct {
	ID        int
	Name      string
	Email     string
	Password  string
	Phone     string
	Photo     string
	Address   string
	Lat       string
	Lng       string
	IsVerified bool
	RoleName  string
}