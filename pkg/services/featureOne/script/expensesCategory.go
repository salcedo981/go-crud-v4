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

func CreateCategory(req *mdlFeatureOne.CreateCategoryRequest) (*mdlFeatureOne.CategoryResponse, error) {
	db := config.DBConnList[0]

	var category mdlFeatureOne.CategoryResponse

	// Insert the new category and return it
	err := db.Raw(`
		INSERT INTO expense_categories (name, description)
		VALUES (?, ?)
		RETURNING id, name, description, created_at
	`, req.Name, req.Description).Scan(&category).Error

	if err != nil {
		log.Printf("[CreateCategory] Error creating category '%s': %v", req.Name, err)
		return nil, err
	}

	log.Printf("[CreateCategory] Success - ID: %d, Name: %s", category.ID, category.Name)
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
