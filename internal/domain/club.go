package domain

import "time"

type ClubStatus string

const (
	ClubStatusPending   ClubStatus = "PENDING"
	ClubStatusActive    ClubStatus = "ACTIVE"
	ClubStatusSuspended ClubStatus = "SUSPENDED"
	ClubStatusArchived  ClubStatus = "ARCHIVED"
)

type Club struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name" binding:"required"`
	Slug               string     `json:"slug" binding:"required"`
	Description        string     `json:"description"`
	Category           string     `json:"category"`
	VerificationStatus string     `json:"verification_status"`
	Status             ClubStatus `json:"status"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type ClubMembershipRole string

const (
	ClubRoleOwner   ClubMembershipRole = "OWNER"
	ClubRoleAdmin   ClubMembershipRole = "ADMIN"
	ClubRoleOfficer ClubMembershipRole = "OFFICER"
	ClubRoleMember  ClubMembershipRole = "MEMBER"
)

type ClubMembership struct {
	ID        string             `json:"id"`
	ClubID    string             `json:"club_id"`
	UserID    string             `json:"user_id"`
	Role      ClubMembershipRole `json:"role"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
