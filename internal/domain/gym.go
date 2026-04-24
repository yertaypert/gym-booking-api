package domain

type Gym struct {
	ID          int    `json:"id"`
	OwnerID     int    `json:"owner_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}
