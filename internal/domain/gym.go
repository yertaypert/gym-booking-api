package domain

type Gym struct {
	ID          int    `json:"id"`
	OwnerID     int    `json:"owner_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type GymWithTrainers struct {
	GymID    int           `json:"gym_id"`
	GymName  string        `json:"gym_name"`
	Trainers []TrainerInfo `json:"trainers"`
}
