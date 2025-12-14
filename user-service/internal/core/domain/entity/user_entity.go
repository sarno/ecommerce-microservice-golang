package entity

type UserEntity struct {
	ID         int
	Name       string
	Email      string
	Password   string
	Phone      string
	Photo      string
	Address    string
	Lat        string
	Lng        string
	IsVerified bool
	RoleName   string
	Token      string
	RoleID     int
}

type QueryStringCustomer struct {
	Search    string
	Page      int
	Limit     int
	OrderBy   string
	OrderType string
}
