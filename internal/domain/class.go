package domain

type Class struct {
	ID          int    `json:"id"`
	GymID       int    `json:"gym_id"`
	Name        string `json:"name"`
	MaxCapacity int    `json:"max_capacity"`
}

type SessionWithGym struct {
	Session
	GymName               string `json:"gym_name"`
	GymAddress            string `json:"gym_address"`
	ClassName             string `json:"class_name"`
	TrainerName           string `json:"trainer_name,omitempty"`
	TrainerSpecialization string `json:"trainer_specialization,omitempty"`
}

type GymWithClass struct {
	GymID      int    `json:"gym_id"`
	GymName    string `json:"gym_name"`
	GymAddress string `json:"gym_address"`
	ClassID    int    `json:"class_id"`
}
