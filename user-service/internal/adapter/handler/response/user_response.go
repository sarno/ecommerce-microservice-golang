package response

type SignResponse struct {
	AccessToken  string `json:"access_token"`
	Role         string `json:"role"`
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Photo        string `json:"photo"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	Lat          string `json:"lat"`
	Lng          string `json:"lng"`
}