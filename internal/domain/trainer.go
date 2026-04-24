package domain

type Trainer struct {
	ID             int     `json:"id"`
	UserID         int     `json:"name"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extraFee"`
}
