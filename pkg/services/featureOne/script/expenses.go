package scpFeatureOne

import (
	"encoding/json"
	"go_template_v3/pkg/config"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"

	"log"
)

// ============================================
// EXPENSE OPERATIONS
// ============================================

// CreateExpense inserts a new expense into the database
func CreateExpense(userID int, req *mdlFeatureOne.CreateExpenseRequest) (*mdlFeatureOne.ExpenseResponse, error) {
	var expense mdlFeatureOne.ExpenseResponse
	var jsonResult string

	err := config.DBConnList[0].Debug().Raw(
		`SELECT * FROM create_expense($1, $2, $3, $4, $5, $6, $7)`,
		userID,
		req.Title,
		req.Amount,
		req.CategoryID,
		req.Date,
		req.Notes,
		req.ImageURL,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[CreateExpense] Error for user %d: %v", userID, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(jsonResult), &expense); err != nil {
		log.Printf("[CreateExpense] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[CreateExpense] Success - ExpenseID: %d, UserID: %d, Title: %s, Amount: %.2f",
		expense.ID, userID, expense.Title, expense.Amount)
	return &expense, nil
}

// GetExpenses retrieves expenses with filters
func GetExpenses(userID int, filters *mdlFeatureOne.ExpenseFilters) (*mdlFeatureOne.ExpenseListResponse, error) {
	// Build filters JSON
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		log.Printf("[GetExpenses] Error marshaling filters: %v", err)
		return nil, err
	}

	var resultJSON string
	err = config.DBConnList[0].Raw(
		`SELECT get_expenses($1, $2::jsonb)`,
		userID,
		string(filtersJSON),
	).Scan(&resultJSON).Error

	if err != nil {
		log.Printf("[GetExpenses] Error for user %d: %v", userID, err)
		return nil, err
	}

	// Parse JSON result
	var result mdlFeatureOne.ExpenseListResponse
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		log.Printf("[GetExpenses] Error parsing result: %v", err)
		return nil, err
	}

	log.Printf("[GetExpenses] Success - UserID: %d, Count: %d, Total: %d",
		userID, len(result.Expenses), result.Pagination.Total)
	return &result, nil
}

// ExpenseExists checks if an expense exists for a user
func ExpenseExists(userID, expenseID int) bool {
	var exists bool

	err := config.DBConnList[0].Raw(
		`SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL)`,
		expenseID,
		userID,
	).Scan(&exists).Error

	if err != nil {
		log.Printf("[ExpenseExists] Error checking expense %d for user %d: %v", expenseID, userID, err)
		return false
	}

	return exists
}

// GetExpenseByID retrieves a single expense by ID
func GetExpenseByID(userID, expenseID int) (*mdlFeatureOne.ExpenseResponse, error) {
	var expense mdlFeatureOne.ExpenseResponse
	var jsonResult string

	err := config.DBConnList[0].Raw(
		`SELECT get_expense_by_id($1, $2)`,
		userID,
		expenseID,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[GetExpenseByID] Error for user %d, expense %d: %v", userID, expenseID, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(jsonResult), &expense); err != nil {
		log.Printf("[GetExpenseByID] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[GetExpenseByID] Success - ExpenseID: %d, UserID: %d, Title: %s",
		expense.ID, userID, expense.Title)
	return &expense, nil
}

// UpdateExpense updates an existing expense
func UpdateExpense(userID, expenseID int, req *mdlFeatureOne.UpdateExpenseRequest) (*mdlFeatureOne.ExpenseResponse, error) {
	var expense mdlFeatureOne.ExpenseResponse
	var jsonResult string

	err := config.DBConnList[0].Raw(
		`SELECT * FROM update_expense($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID,
		expenseID,
		req.Title,
		req.Amount,
		req.CategoryID,
		req.Date,
		req.Notes,
		req.ImageURL,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[UpdateExpense] Error for user %d, expense %d: %v", userID, expenseID, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(jsonResult), &expense); err != nil {
		log.Printf("[UpdateExpense] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[UpdateExpense] Success - ExpenseID: %d, UserID: %d, Title: %s",
		expense.ID, userID, expense.Title)
	return &expense, nil
}

// DeleteExpense soft deletes an expense
func DeleteExpense(userID, expenseID int) (*mdlFeatureOne.DeleteExpenseResult, error) {
	var result mdlFeatureOne.DeleteExpenseResult

	err := config.DBConnList[0].Raw(
		`SELECT * FROM delete_expense($1, $2)`,
		userID,
		expenseID,
	).Scan(&result).Error

	if err != nil {
		log.Printf("[DeleteExpense] Error for user %d, expense %d: %v", userID, expenseID, err)
		return nil, err
	}

	if !result.Deleted {
		log.Printf("[DeleteExpense] Expense not found - UserID: %d, ExpenseID: %d", userID, expenseID)
		return &result, nil
	}

	log.Printf("[DeleteExpense] Success - ExpenseID: %d, UserID: %d, Had Image: %v",
		expenseID, userID, result.ImageURL != nil)
	return &result, nil
}
