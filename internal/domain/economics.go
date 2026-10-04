package domain

import "time"

type EconDataset struct {
	ID              string                 `json:"id"`
	SourceName      string                 `json:"source_name" binding:"required"`
	Publisher       string                 `json:"publisher"`
	Version         string                 `json:"version" binding:"required"`
	Methodology     string                 `json:"methodology"`
	LicenseInfo     string                 `json:"license_info"`
	GeographicScope string                 `json:"geographic_scope"`
	UpdateFrequency string                 `json:"update_frequency"`
	Provenance      map[string]interface{} `json:"provenance"`
	Status          string                 `json:"status"`
}

type EconIndicator struct {
	ID               string                 `json:"id"`
	DatasetID        string                 `json:"dataset_id" binding:"required"`
	IndicatorType    string                 `json:"indicator_type" binding:"required"`
	Region           string                 `json:"region"`
	TimePeriod       *time.Time             `json:"time_period" binding:"required"`
	Value            float64                `json:"value"`
	UncertaintyRange map[string]interface{} `json:"uncertainty_range"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type RegionalSkillBalance struct {
	ID              string     `json:"id"`
	Region          string     `json:"region" binding:"required"`
	SkillName       string     `json:"skill_name" binding:"required"`
	WorkforceDemand float64    `json:"workforce_demand"`
	EducationSupply float64    `json:"education_supply"`
	BalanceStatus   string     `json:"balance_status"`
	Confidence      float64    `json:"confidence"`
	LastComputedAt  *time.Time `json:"last_computed_at,omitempty"`
}

type TechTrend struct {
	ID                         string                 `json:"id"`
	TechnologyName             string                 `json:"technology_name" binding:"required"`
	MaturityLevel              string                 `json:"maturity_level"`
	AffectedOccupations        []interface{}          `json:"affected_occupations"`
	TaskAutomationRisk         map[string]interface{} `json:"task_automation_risk"`
	HumanAugmentationPotential map[string]interface{} `json:"human_augmentation_potential"`
	Evidence                   []interface{}          `json:"evidence"`
}

type PolicyScenario struct {
	ID                 string                 `json:"id"`
	OwnerID            string                 `json:"owner_id"`
	Title              string                 `json:"title" binding:"required"`
	Assumptions        map[string]interface{} `json:"assumptions" binding:"required"`
	ModelVersion       string                 `json:"model_version"`
	Status             string                 `json:"status"`
	Outputs            map[string]interface{} `json:"outputs"`
	PolicyImplications map[string]interface{} `json:"policy_implications"`
	CreatedAt          *time.Time             `json:"created_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
}

type FundingOpportunity struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title" binding:"required"`
	FundingType      string                 `json:"funding_type"`
	ProviderName     string                 `json:"provider_name"`
	EligibilityRules map[string]interface{} `json:"eligibility_rules"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
	Status           string                 `json:"status"`
}
