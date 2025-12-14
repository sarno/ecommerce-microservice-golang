package request

// UpdateCustomerRequest defines the request body for updating a customer.
// Password fields are optional. If a new password is provided,
// it must be at least 8 characters and match the confirmation.
type UpdateCustomerRequest struct {
	Name                 string  `json:"name" validate:"required"`
	Email                string  `json:"email" validate:"email,required"`
	Password             string  `json:"password" validate:"omitempty,min=8"`
	PasswordConfirmation string  `json:"password_confirmation" validate:"omitempty,eqfield=Password"`
	Phone                string  `json:"phone" validate:"required,number"`
	Address              string  `json:"address"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	Photo                string  `json:"photo"`
	RoleID               int     `json:"role_id" validate:"required"`
}
