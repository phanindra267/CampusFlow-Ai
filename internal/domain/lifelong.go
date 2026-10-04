package domain

import "time"

type LifelongProfile struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id" binding:"required"`
	CareerGoals     []interface{}          `json:"career_goals"`
	VerifiedSkills  []interface{}          `json:"verified_skills"`
	LearningHistory []interface{}          `json:"learning_history"`
	Portfolio       []interface{}          `json:"portfolio"`
	PrivacyControls map[string]interface{} `json:"privacy_controls"`
}

type WorkforceSkill struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name" binding:"required"`
	Category         string                 `json:"category"`
	IndustryMappings map[string]interface{} `json:"industry_mappings"`
	RelatedSkills    []interface{}          `json:"related_skills"`
	Version          string                 `json:"version"`
	Status           string                 `json:"status"`
}

type JobRole struct {
	ID                     string                 `json:"id"`
	Title                  string                 `json:"title" binding:"required"`
	EmployerID             string                 `json:"employer_id" binding:"required"`
	Responsibilities       []interface{}          `json:"responsibilities"`
	RequiredSkills         []interface{}          `json:"required_skills"`
	PreferredSkills        []interface{}          `json:"preferred_skills"`
	ExperienceRequirements map[string]interface{} `json:"experience_requirements"`
	EducationRequirements  map[string]interface{} `json:"education_requirements"`
	LocationPreferences    map[string]interface{} `json:"location_preferences"`
	Status                 string                 `json:"status"`
}

type CareerPath struct {
	ID                   string                 `json:"id"`
	UserID               string                 `json:"user_id" binding:"required"`
	TargetRoleID         string                 `json:"target_role_id" binding:"required"`
	SkillGaps            []interface{}          `json:"skill_gaps"`
	ClosurePlan          map[string]interface{} `json:"closure_plan"`
	TransitionReadiness  float64                `json:"transition_readiness"`
	EstimatedEffortHours int                    `json:"estimated_effort_hours"`
	Status               string                 `json:"status"`
}

type VerifiedExperience struct {
	ID                 string        `json:"id"`
	UserID             string        `json:"user_id" binding:"required"`
	VerifierID         string        `json:"verifier_id" binding:"required"`
	ExperienceType     string        `json:"experience_type" binding:"required"`
	Description        string        `json:"description"`
	DemonstratedSkills []interface{} `json:"demonstrated_skills"`
	VerificationStatus string        `json:"verification_status"`
	VerifiedAt         *time.Time    `json:"verified_at,omitempty"`
}

type TalentMatch struct {
	ID               string                 `json:"id"`
	JobRoleID        string                 `json:"job_role_id" binding:"required"`
	UserID           string                 `json:"user_id" binding:"required"`
	MatchScore       float64                `json:"match_score"`
	MatchExplanation map[string]interface{} `json:"match_explanation"`
	EmployerStatus   string                 `json:"employer_status"`
	CandidateStatus  string                 `json:"candidate_status"`
}
