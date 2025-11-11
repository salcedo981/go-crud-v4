package ctrFeatureOne

import (
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	scpFeatureOne "go_template_v3/pkg/services/featureOne/script"
	"net/http"
	"strings"

	v1 "github.com/FDSAP-Git-Org/hephaestus/helper/v1"
	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	"github.com/gofiber/fiber/v3"
)

// ============================================
// CATEGORY ENDPOINTS
// ============================================

// CreateCategory creates a new expense category
func CreateCategory(c fiber.Ctx) error {
	var req mdlFeatureOne.CreateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Name is required", nil, http.StatusBadRequest)
	}

	// Create category
	category, err := scpFeatureOne.CreateCategory(&req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Category name already exists", err, http.StatusBadRequest)
		}
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create category", err, http.StatusInternalServerError)
	}

	// Map to response
	response := mdlFeatureOne.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_201,
		"Category created successfully", response, http.StatusCreated)
}

// GetCategories retrieves all categories with pagination
func GetCategories(c fiber.Ctx) error {
	// Parse query parameters
	filters := &mdlFeatureOne.CategoryFilters{
		Limit:  getQueryIntDefault(c, "limit", 50),
		Offset: getQueryIntDefault(c, "offset", 0),
	}

	// Validate limit
	if filters.Limit < 1 || filters.Limit > 100 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Limit must be between 1 and 100", nil, http.StatusBadRequest)
	}

	// Get categories
	result, err := scpFeatureOne.GetCategories(filters)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to retrieve categories", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Categories retrieved successfully", result, http.StatusOK)
}
