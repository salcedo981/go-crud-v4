package scpFeatureOne

import (
	"encoding/json"
	"fmt"
	"go_template_v3/pkg/config"
	hlpFeatureOne "go_template_v3/pkg/services/featureOne/helper"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	"log"
	"time"
)

// ============================================
// BATCH JOB OPERATIONS
// ============================================

// BatchJobExists checks if a batch job exists for a user
func BatchJobExists(userID, jobID int) bool {
	var exists bool

	err := config.DBConnList[0].Raw(
		`SELECT EXISTS(SELECT 1 FROM batch_jobs WHERE id = $1 AND user_id = $2)`,
		jobID,
		userID,
	).Scan(&exists).Error

	if err != nil {
		log.Printf("[BatchJobExists] Error checking job %d for user %d: %v", jobID, userID, err)
		return false
	}

	return exists
}

// CreateBatchJob creates a new batch job record
func CreateBatchJob(userID int, jobType string, totalItems int) (int, error) {
	var jobID int

	err := config.DBConnList[0].Raw(
		`SELECT create_batch_job($1, $2, $3)`,
		userID,
		jobType,
		totalItems,
	).Scan(&jobID).Error

	if err != nil {
		log.Printf("[CreateBatchJob] Error for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("[CreateBatchJob] Success - JobID: %d, UserID: %d, Type: %s, Items: %d",
		jobID, userID, jobType, totalItems)
	return jobID, nil
}

// UpdateBatchJob updates batch job progress
func UpdateBatchJob(jobID int, status string, processed, successful, failed int, results interface{}) error {
	// Marshal results to JSON
	var resultsJSON *string
	if results != nil {
		jsonBytes, err := json.Marshal(results)
		if err != nil {
			log.Printf("[UpdateBatchJob] Error marshaling results: %v", err)
			return err
		}
		jsonStr := string(jsonBytes)
		resultsJSON = &jsonStr
	}

	var success bool
	err := config.DBConnList[0].Raw(
		`SELECT update_batch_job($1, $2, $3, $4, $5, $6)`,
		jobID,
		status,
		processed,
		successful,
		failed,
		resultsJSON,
	).Scan(&success).Error

	if err != nil {
		log.Printf("[UpdateBatchJob] Error updating job %d: %v", jobID, err)
		return err
	}

	log.Printf("[UpdateBatchJob] JobID: %d, Status: %s, Processed: %d/%d, Success: %d, Failed: %d",
		jobID, status, processed, processed, successful, failed)
	return nil
}

// GetBatchJob retrieves batch job status
func GetBatchJob(userID, jobID int) (*mdlFeatureOne.BatchJobResponse, error) {
	var resultJSON string

	err := config.DBConnList[0].Raw(
		`SELECT get_batch_job($1, $2)`,
		userID,
		jobID,
	).Scan(&resultJSON).Error

	if err != nil {
		log.Printf("[GetBatchJob] Error for user %d, job %d: %v", userID, jobID, err)
		return nil, err
	}

	// Parse JSON result
	var result mdlFeatureOne.BatchJobResponse
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		log.Printf("[GetBatchJob] Error parsing result: %v", err)
		return nil, err
	}

	log.Printf("[GetBatchJob] Success - JobID: %d, Status: %s, Progress: %d/%d",
		result.JobID, result.Status, result.ProcessedItems, result.TotalItems)
	return &result, nil
}

// ============================================
// BATCH PROCESSING OPERATIONS
// ============================================

// ProcessBatchUpdate processes batch expense updates asynchronously
func ProcessBatchUpdate(jobID, userID int, updates []mdlFeatureOne.BatchUpdateItem) {
	log.Printf("[ProcessBatchUpdate] Starting - JobID: %d, Items: %d", jobID, len(updates))

	// Update status to processing
	UpdateBatchJob(jobID, "processing", 0, 0, 0, nil)

	var results []mdlFeatureOne.BatchUpdateResultItem
	successCount := 0
	failCount := 0

	for i, update := range updates {
		// Add delay to prevent overwhelming database
		time.Sleep(30 * time.Second)

		// Build update request
		req := &mdlFeatureOne.UpdateExpenseRequest{
			Title:      update.Title,
			Amount:     update.Amount,
			CategoryID: update.CategoryID,
			Date:       update.Date,
			Notes:      update.Notes,
		}

		if !ExpenseExists(userID, update.ExpenseID) {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:     i,
				ExpenseID: update.ExpenseID,
				Success:   false,
				Message:   "Expense not found",
			})
			continue
		}

		// Attempt update
		_, err := UpdateExpense(userID, update.ExpenseID, req)

		if err != nil {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:     i,
				ExpenseID: update.ExpenseID,
				Success:   false,
				Message:   "Failed to update expense",
			})
			log.Printf("[ProcessBatchUpdate] Failed - Item %d, ExpenseID: %d", i, update.ExpenseID)
		} else {
			successCount++
			log.Printf("[ProcessBatchUpdate] Success - Item %d, ExpenseID: %d", i, update.ExpenseID)
		}

		// Update progress
		UpdateBatchJob(jobID, "processing", i+1, successCount, failCount, results)
	}

	// Mark as completed
	finalStatus := "completed"
	if failCount == len(updates) {
		finalStatus = "failed"
	}

	UpdateBatchJob(jobID, finalStatus, len(updates), successCount, failCount, results)
	log.Printf("[ProcessBatchUpdate] Completed - JobID: %d, Success: %d, Failed: %d",
		jobID, successCount, failCount)
}

// ProcessBatchUpload processes batch expense uploads from CSV asynchronously
func ProcessBatchUpload(jobID, userID int, expenses []mdlFeatureOne.CSVExpenseRow) {
	log.Printf("[ProcessBatchUpload] Starting - JobID: %d, Items: %d", jobID, len(expenses))

	// Update status to processing
	UpdateBatchJob(jobID, "processing", 0, 0, 0, nil)

	var results []mdlFeatureOne.BatchUpdateResultItem
	successCount := 0
	failCount := 0

	for i, expense := range expenses {
		// Add delay to prevent overwhelming database
		time.Sleep(500 * time.Millisecond)

		// Build create request
		req := &mdlFeatureOne.CreateExpenseRequest{
			Title:      expense.Title,
			Amount:     expense.Amount,
			CategoryID: expense.CategoryID,
			Date:       expense.Date,
			Notes:      expense.Notes,
		}

		// Validate before inserting
		if err := hlpFeatureOne.ValidateCreateExpense(req); err != nil {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:   i,
				Success: false,
				Message: fmt.Sprintf("Validation failed: %s", err.Error()),
			})
			log.Printf("[ProcessBatchUpload] Validation failed - Item %d: %s", i, err.Error())
			UpdateBatchJob(jobID, "processing", i+1, successCount, failCount, results)
			continue
		}

		// Attempt create
		created, err := CreateExpense(userID, req)

		if err != nil {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:   i,
				Success: false,
				Message: "Failed to create expense",
			})
			log.Printf("[ProcessBatchUpload] Failed - Item %d", i)
		} else {
			successCount++
			log.Printf("[ProcessBatchUpload] Success - Item %d, ExpenseID: %d", i, created.ID)
		}

		// Update progress
		UpdateBatchJob(jobID, "processing", i+1, successCount, failCount, results)
	}

	// Mark as completed
	finalStatus := "completed"
	if failCount == len(expenses) {
		finalStatus = "failed"
	}

	UpdateBatchJob(jobID, finalStatus, len(expenses), successCount, failCount, results)
	log.Printf("[ProcessBatchUpload] Completed - JobID: %d, Success: %d, Failed: %d",
		jobID, successCount, failCount)
}
