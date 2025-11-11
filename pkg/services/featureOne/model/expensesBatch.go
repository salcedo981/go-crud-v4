package mdlFeatureOne

// ============================================
// BATCH REQUEST STRUCTS
// ============================================

type BatchUpdateItem struct {
	ExpenseID  int      `json:"expenseId"`
	Title      *string  `json:"title"`
	Amount     *float64 `json:"amount"`
	CategoryID *int     `json:"categoryId"`
	Date       *string  `json:"date"`
	Notes      *string  `json:"notes"`
}

type BatchUpdateRequest struct {
	Updates []BatchUpdateItem `json:"updates"`
}

type CSVExpenseRow struct {
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
	CategoryID *int    `json:"categoryId"`
	Date       string  `json:"date"`
	Notes      *string `json:"notes"`
}

type BatchUploadRequest struct {
	Expenses []CSVExpenseRow `json:"expenses"`
}

// ============================================
// BATCH RESPONSE STRUCTS
// ============================================

type BatchUpdateResultItem struct {
	Index     int    `json:"index"`
	ExpenseID int    `json:"expenseId"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
}

type BatchUpdateResponse struct {
	Results    []BatchUpdateResultItem `json:"results"`
	Total      int                     `json:"total"`
	Successful int                     `json:"successful"`
	Failed     int                     `json:"failed"`
}

type BatchJobResponse struct {
	JobID           int         `json:"jobId"`
	UserID          int         `json:"userId"`
	JobType         string      `json:"jobType"`
	Status          string      `json:"status"`
	TotalItems      int         `json:"totalItems"`
	ProcessedItems  int         `json:"processedItems"`
	SuccessfulItems int         `json:"successfulItems"`
	FailedItems     int         `json:"failedItems"`
	Results         interface{} `json:"results"`
	CreatedAt       string      `json:"createdAt"`
	UpdatedAt       string      `json:"updatedAt"`
	CompletedAt     *string     `json:"completedAt"`
}

type BatchJobCreatedResponse struct {
	JobID      int    `json:"jobId"`
	TotalItems int    `json:"totalItems"`
	Status     string `json:"status"`
}

// ============================================
// BATCH ENTITY STRUCTS (DB)
// ============================================

type BatchJobEntity struct {
	ID              int     `db:"id"`
	UserID          int     `db:"user_id"`
	JobType         string  `db:"job_type"`
	Status          string  `db:"status"`
	TotalItems      int     `db:"total_items"`
	ProcessedItems  int     `db:"processed_items"`
	SuccessfulItems int     `db:"successful_items"`
	FailedItems     int     `db:"failed_items"`
	Results         *string `db:"results"`
	CreatedAt       string  `db:"created_at"`
	UpdatedAt       string  `db:"updated_at"`
	CompletedAt     *string `db:"completed_at"`
}

// ============================================
// HELPER STRUCTS
// ============================================

// type BatchProcessResult struct {
// 	Index     int    `json:"index"`
// 	ExpenseID *int   `json:"expenseId,omitempty"`
// 	Status    string `json:"status"`
// 	Message   string `json:"message"`
// }
