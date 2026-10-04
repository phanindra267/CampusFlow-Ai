package domain

import "time"

type StudentProfile struct {
	ID                  string                 `json:"id"`
	TenantID            string                 `json:"tenant_id"`
	StudentID           string                 `json:"student_id" binding:"required"`
	AcademicStanding    string                 `json:"academic_standing"`
	EnrolledCourses     []interface{}          `json:"enrolled_courses"`
	MasteredSkills      map[string]interface{} `json:"mastered_skills"`
	LearningPreferences map[string]interface{} `json:"learning_preferences"`
	PrivacyControls     map[string]interface{} `json:"privacy_controls"`
}

type EduLearningPath struct {
	ID                    string        `json:"id"`
	StudentProfileID      string        `json:"student_profile_id" binding:"required"`
	Objective             string        `json:"objective" binding:"required"`
	TargetCompetency      string        `json:"target_competency"`
	Topics                []interface{} `json:"topics"`
	Resources             []interface{} `json:"resources"`
	ExpectedDurationHours int           `json:"expected_duration_hours"`
	Status                string        `json:"status"`
}

type EduIntervention struct {
	ID                  string                 `json:"id"`
	StudentProfileID    string                 `json:"student_profile_id" binding:"required"`
	DetectedChallenge   string                 `json:"detected_challenge" binding:"required"`
	RecommendedAction   string                 `json:"recommended_action"`
	Evidence            map[string]interface{} `json:"evidence"`
	AdvisorReviewStatus string                 `json:"advisor_review_status"`
	StudentFeedback     map[string]interface{} `json:"student_feedback"`
}

type EduAssessment struct {
	ID                 string                 `json:"id"`
	CourseID           string                 `json:"course_id" binding:"required"`
	Title              string                 `json:"title" binding:"required"`
	AssessmentMode     string                 `json:"assessment_mode"`
	AIAssistancePolicy string                 `json:"ai_assistance_policy"`
	Questions          []interface{}          `json:"questions"`
	ItemAnalytics      map[string]interface{} `json:"item_analytics"`
}

type EduOpportunity struct {
	ID              string        `json:"id"`
	TenantID        string        `json:"tenant_id"`
	OpportunityType string        `json:"opportunity_type"`
	Title           string        `json:"title" binding:"required"`
	RequiredSkills  []interface{} `json:"required_skills"`
	TargetSkillGaps []interface{} `json:"target_skill_gaps"`
	Status          string        `json:"status"`
	CreatedAt       *time.Time    `json:"created_at,omitempty"`
}
