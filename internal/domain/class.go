package domain

type Class struct {
	ID          int    `json:"id"`
	GymID       int    `json:"gym_id"`
	Name        string `json:"name"`
	MaxCapacity int    `json:"max_capacity"`
}
