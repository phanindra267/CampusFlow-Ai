package domain

import "time"

// People directory. The directory exists so a member can find the person who can
// help: a faculty member who supervises in an area they want to research, a
// mentor for a project, a coordinator for a scheme. It is a discovery surface,
// not a staff record, and an entry may exist for someone without a platform
// account.
const (
	PersonRoleFaculty     = "FACULTY"
	PersonRoleResearcher  = "RESEARCHER"
	PersonRoleSupervisor  = "SUPERVISOR"
	PersonRoleMentor      = "MENTOR"
	PersonRoleCoordinator = "COORDINATOR"
)

// Person is one directory entry.
type Person struct {
	ID                string    `json:"id"`
	UserID            *string   `json:"user_id,omitempty"`
	DisplayName       string    `json:"display_name"`
	Role              string    `json:"role"`
	School            string    `json:"school,omitempty"`
	Designation       string    `json:"designation,omitempty"`
	Bio               string    `json:"bio,omitempty"`
	Email             string    `json:"email,omitempty"`
	OfficeLocation    string    `json:"office_location,omitempty"`
	ProfileURL        string    `json:"profile_url,omitempty"`
	ResearchInterests []string  `json:"research_interests,omitempty"`
	AcceptingStudents bool      `json:"accepting_students"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreatePersonRequest adds a directory entry. Administrators own this surface
// because it names real staff.
type CreatePersonRequest struct {
	DisplayName       string   `json:"display_name" binding:"required,min=2,max=255"`
	Role              string   `json:"role" binding:"required,oneof=FACULTY RESEARCHER SUPERVISOR MENTOR COORDINATOR"`
	School            string   `json:"school" binding:"omitempty,max=255"`
	Designation       string   `json:"designation" binding:"omitempty,max=255"`
	Bio               string   `json:"bio" binding:"omitempty,max=5000"`
	Email             string   `json:"email" binding:"omitempty,email,max=255"`
	OfficeLocation    string   `json:"office_location" binding:"omitempty,max=255"`
	ProfileURL        string   `json:"profile_url" binding:"omitempty,url,max=512"`
	ResearchInterests []string `json:"research_interests" binding:"omitempty,max=20,dive,max=150"`
	AcceptingStudents bool     `json:"accepting_students"`
	UserID            *string  `json:"user_id,omitempty"`
}

// UpdatePersonRequest patches a directory entry.
type UpdatePersonRequest struct {
	DisplayName       *string   `json:"display_name,omitempty" binding:"omitempty,min=2,max=255"`
	Role              *string   `json:"role,omitempty" binding:"omitempty,oneof=FACULTY RESEARCHER SUPERVISOR MENTOR COORDINATOR"`
	School            *string   `json:"school,omitempty" binding:"omitempty,max=255"`
	Designation       *string   `json:"designation,omitempty" binding:"omitempty,max=255"`
	Bio               *string   `json:"bio,omitempty" binding:"omitempty,max=5000"`
	Email             *string   `json:"email,omitempty" binding:"omitempty,email,max=255"`
	OfficeLocation    *string   `json:"office_location,omitempty" binding:"omitempty,max=255"`
	ProfileURL        *string   `json:"profile_url,omitempty" binding:"omitempty,url,max=512"`
	ResearchInterests *[]string `json:"research_interests,omitempty" binding:"omitempty,max=20,dive,max=150"`
	AcceptingStudents *bool     `json:"accepting_students,omitempty"`
	Status            *string   `json:"status,omitempty" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}

// ResearchInterest is one topic on a member's own research profile. Keeping it
// separate from general career interests lets research matching read one list.
type ResearchInterest struct {
	UserID    string    `json:"user_id"`
	Interest  string    `json:"interest"`
	CreatedAt time.Time `json:"created_at"`
}

// AddResearchInterestRequest adds a research interest to the signed-in member.
type AddResearchInterestRequest struct {
	Interest string `json:"interest" binding:"required,min=3,max=150"`
}

// MemberMatch is a fellow member found by shared research interest. It is
// deliberately thinner than a directory entry: a member only sees the context a
// collaborator needs, not a staff record.
type MemberMatch struct {
	UserID              string   `json:"user_id"`
	DisplayName         string   `json:"display_name"`
	Headline            string   `json:"headline,omitempty"`
	Program             string   `json:"program,omitempty"`
	Branch              string   `json:"branch,omitempty"`
	YearOfStudy         *int     `json:"year_of_study,omitempty"`
	AvatarURL           string   `json:"avatar_url,omitempty"`
	SharedInterests     []string `json:"shared_interests"`
	OpenToOpportunities bool     `json:"open_to_opportunities"`
}
