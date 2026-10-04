package domain

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"` // student, organizer, admin, superadmin
	Description string `json:"description"`
}
