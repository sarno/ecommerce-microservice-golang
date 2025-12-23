package entity

type CustomerResponseEntity struct {
	RoleID  int    `json:"role_id"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Lat     string `json:"lat"`
	Lng     string `json:"lng"`
	Address string `json:"address"`
	Photo   string `json:"photo"`
}