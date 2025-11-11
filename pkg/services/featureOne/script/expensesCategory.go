package scpFeatureOne

import (
	"encoding/json"
	"go_template_v3/pkg/config"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"

	"log"
)

// ============================================
// CATEGORY OPERATIONS
// ============================================

// CreateCategory creates a new expense category
func CreateCategory(req *mdlFeatureOne.CreateCategoryRequest) (*mdlFeatureOne.CategoryEntity, error) {
	var category mdlFeatureOne.CategoryEntity

	err := config.DBConnList[0].Raw(
		`SELECT * FROM create_category($1, $2)`,
		req.Name,
		req.Description,
	).Scan(&category).Error

	if err != nil {
		log.Printf("[CreateCategory] Error creating category '%s': %v", req.Name, err)
		return nil, err
	}

	log.Printf("[CreateCategory] Success - CategoryID: %d, Name: %s", category.ID, category.Name)
	return &category, nil
}

// GetCategories retrieves all categories with pagination
func GetCategories(filters *mdlFeatureOne.CategoryFilters) (*mdlFeatureOne.CategoryListResponse, error) {
	var resultJSON string

	err := config.DBConnList[0].Raw(
		`SELECT get_categories($1, $2)`,
		filters.Limit,
		filters.Offset,
	).Scan(&resultJSON).Error

	if err != nil {
		log.Printf("[GetCategories] Error: %v", err)
		return nil, err
	}

	// Parse JSON result
	var result mdlFeatureOne.CategoryListResponse
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		log.Printf("[GetCategories] Error parsing result: %v", err)
		return nil, err
	}

	log.Printf("[GetCategories] Success - Count: %d, Total: %d",
		len(result.Categories), result.Pagination.Total)
	return &result, nil
}
