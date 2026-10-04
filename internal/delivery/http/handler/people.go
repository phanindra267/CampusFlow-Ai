package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

// PeopleHandler serves the campus people directory and the member's own research
// profile. The directory is how a member finds a supervisor, mentor or
// coordinator; the research profile is how they say what they want to work on.
type PeopleHandler struct {
	db *postgres.PeopleRepository
}

func NewPeopleHandler(db *postgres.PeopleRepository) *PeopleHandler {
	return &PeopleHandler{db: db}
}

// ListPeople searches the directory.
func (h *PeopleHandler) ListPeople(c *gin.Context) {
	people, err := h.db.ListPeople(c.Request.Context(), c.Query("q"), c.Query("role"),
		c.Query("school"), c.Query("interest"), c.Query("accepting") == "true",
		pageSize(c.Query("limit")), parsePositiveInt(c.Query("offset"), 0))
	if err != nil {
		writeRepoError(c, err, "PERSON_NOT_FOUND", "PEOPLE_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "People retrieved", gin.H{
		"people": people,
		"count":  len(people),
	})
}

// GetPerson returns one directory entry.
func (h *PeopleHandler) GetPerson(c *gin.Context) {
	person, err := h.db.GetPerson(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRepoError(c, err, "PERSON_NOT_FOUND", "PERSON_FETCH_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Person retrieved", person)
}

// ListSchools returns the schools represented in the directory so the browse
// filter offers real values.
func (h *PeopleHandler) ListSchools(c *gin.Context) {
	schools, err := h.db.ListSchools(c.Request.Context())
	if err != nil {
		writeRepoError(c, err, "SCHOOL_NOT_FOUND", "SCHOOL_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Schools retrieved", gin.H{"schools": schools})
}

// ListResearchInterests returns the topics people in the directory work on.
func (h *PeopleHandler) ListResearchInterests(c *gin.Context) {
	interests, err := h.db.ListResearchInterests(c.Request.Context(), pageSize(c.Query("limit")))
	if err != nil {
		writeRepoError(c, err, "INTEREST_NOT_FOUND", "INTEREST_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Research interests retrieved", gin.H{
		"interests": interests,
	})
}

// FindPeopleByInterest is the collaboration lookup: everyone working on a topic,
// optionally only those currently taking students.
func (h *PeopleHandler) FindPeopleByInterest(c *gin.Context) {
	interest := c.Query("interest")
	if interest == "" {
		response.Error(c, http.StatusBadRequest, "INTEREST_REQUIRED", nil)
		return
	}

	people, err := h.db.FindPeopleByInterest(c.Request.Context(), interest,
		c.Query("accepting") == "true", pageSize(c.Query("limit")))
	if err != nil {
		writeRepoError(c, err, "PERSON_NOT_FOUND", "PEOPLE_INTEREST_SEARCH_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "People with this interest retrieved", gin.H{
		"interest": interest,
		"people":   people,
		"count":    len(people),
	})
}

// FindMembersByResearchInterest finds fellow members who listed the same topic, so
// a member can find collaborators rather than only supervisors.
func (h *PeopleHandler) FindMembersByResearchInterest(c *gin.Context) {
	interest := c.Query("interest")
	if interest == "" {
		response.Error(c, http.StatusBadRequest, "INTEREST_REQUIRED", nil)
		return
	}

	members, err := h.db.FindMembersByResearchInterest(c.Request.Context(), middleware.UserID(c),
		interest, c.Query("program"), pageSize(c.Query("limit")))
	if err != nil {
		writeRepoError(c, err, "MEMBER_NOT_FOUND", "MEMBER_INTEREST_SEARCH_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Members with this interest retrieved", gin.H{
		"interest": interest,
		"members":  members,
		"count":    len(members),
	})
}

// CreatePerson adds a directory entry. Administrators own this surface because
// it names real staff, so it is admin-only at the route.
func (h *PeopleHandler) CreatePerson(c *gin.Context) {
	var req domain.CreatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	person, err := h.db.CreatePerson(c.Request.Context(), &req)
	if err != nil {
		writeRepoError(c, err, "PERSON_NOT_FOUND", "PERSON_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Person added to directory", person)
}

func (h *PeopleHandler) UpdatePerson(c *gin.Context) {
	var req domain.UpdatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	person, err := h.db.UpdatePerson(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		writeRepoError(c, err, "PERSON_NOT_FOUND", "PERSON_UPDATE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Person updated", person)
}

// ---------------------------------------------------- member research profile

// ListMyResearchInterests returns the signed-in member's own topics.
func (h *PeopleHandler) ListMyResearchInterests(c *gin.Context) {
	interests, err := h.db.ListMemberResearchInterests(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeRepoError(c, err, "INTEREST_NOT_FOUND", "INTEREST_LIST_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Your research interests retrieved", gin.H{
		"interests": interests,
		"count":     len(interests),
	})
}

// AddMyResearchInterest records a topic the member wants to work on.
func (h *PeopleHandler) AddMyResearchInterest(c *gin.Context) {
	var req domain.AddResearchInterestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	if err := h.db.AddMemberResearchInterest(c.Request.Context(), middleware.UserID(c),
		req.Interest); err != nil {
		writeRepoError(c, err, "INTEREST_NOT_FOUND", "INTEREST_CREATE_FAILED")
		return
	}
	response.Success(c, http.StatusCreated, "Research interest added",
		gin.H{"interest": req.Interest})
}

func (h *PeopleHandler) RemoveMyResearchInterest(c *gin.Context) {
	if err := h.db.RemoveMemberResearchInterest(c.Request.Context(), middleware.UserID(c),
		c.Param("interest")); err != nil {
		writeRepoError(c, err, "INTEREST_NOT_FOUND", "INTEREST_DELETE_FAILED")
		return
	}
	response.Success(c, http.StatusOK, "Research interest removed", nil)
}
