package mdlFeatureOne

// ============================================
// EXPENSE REQUEST STRUCTS
// ============================================

type CreateExpenseRequest struct {
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
	CategoryID *int    `json:"categoryId"`
	Date       string  `json:"date"`
	Notes      *string `json:"notes"`
	ImageURL   *string `json:"imageUrl"`
}

type UpdateExpenseRequest struct {
	Title      *string  `json:"title"`
	Amount     *float64 `json:"amount"`
	CategoryID *int     `json:"categoryId"`
	Date       *string  `json:"date"`
	Notes      *string  `json:"notes"`
	ImageURL   *string  `json:"imageUrl"`
}

type ExpenseFilters struct {
	Title      *string  `json:"title"`
	MinAmount  *float64 `json:"minAmount"`
	MaxAmount  *float64 `json:"maxAmount"`
	CategoryID *int     `json:"categoryId"`
	StartDate  *string  `json:"startDate"`
	EndDate    *string  `json:"endDate"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// ============================================
// EXPENSE RESPONSE STRUCTS
// ============================================

type CategoryInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExpenseResponse struct {
	ID        int           `json:"id"`
	Title     string        `json:"title"`
	Amount    float64       `json:"amount"`
	Category  *CategoryInfo `json:"category"`
	Date      string        `json:"date"`
	Notes     *string       `json:"notes"`
	ImageURL  *string       `json:"imageUrl"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
}

type PaginationResponse struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ExpenseListResponse struct {
	Expenses   []ExpenseResponse  `json:"expenses"`
	Pagination PaginationResponse `json:"pagination"`
}

// ============================================
// EXPENSE ENTITY STRUCTS (DB)
// ============================================

type ExpenseEntity struct {
	ID                  int     `db:"id"`
	UserID              int     `db:"user_id"`
	Title               string  `db:"title"`
	Amount              float64 `db:"amount"`
	CategoryID          *int    `db:"category_id"`
	CategoryName        *string `db:"category_name"`
	CategoryDescription *string `db:"category_description"`
	Date                string  `db:"date"`
	Notes               *string `db:"notes"`
	ImageURL            *string `db:"image_url"`
	CreatedAt           string  `db:"created_at"`
	UpdatedAt           string  `db:"updated_at"`
	DeletedAt           *string `db:"deleted_at"`
}

// ============================================
// HELPER STRUCTS
// ============================================

type DeleteExpenseResult struct {
	Deleted  bool    `db:"deleted"`
	ImageURL *string `db:"image_url"`
}
