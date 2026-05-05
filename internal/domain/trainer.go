package domain

type Trainer struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extraFee"`
}

type TrainerInfo struct {
	UserID         int     `json:"user_id"`
	FullName       string  `json:"full_name"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extra_fee"`
}
