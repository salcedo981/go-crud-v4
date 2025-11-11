package mdlFeatureOne

// ============================================
// CATEGORY REQUEST STRUCTS
// ============================================

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type CategoryFilters struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ============================================
// CATEGORY RESPONSE STRUCTS
// ============================================

type CategoryResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Pagination PaginationResponse `json:"pagination"`
}

// ============================================
// CATEGORY ENTITY STRUCTS (DB)
// ============================================

type CategoryEntity struct {
	ID          int     `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
	CreatedAt   string  `db:"created_at"`
	UpdatedAt   string  `db:"updated_at"`
}
