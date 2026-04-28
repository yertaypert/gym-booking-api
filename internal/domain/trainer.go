package domain

type Trainer struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extraFee"`
}
