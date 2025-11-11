package ctrFeatureOne

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	v1 "github.com/FDSAP-Git-Org/hephaestus/helper/v1"
	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	"github.com/gofiber/fiber/v3"

	"go_template_v3/pkg/global/utils"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	scpFeatureOne "go_template_v3/pkg/services/featureOne/script"
)

// ============================================
// BATCH UPDATE ENDPOINTS
// ============================================

// BatchUpdateExpenses performs synchronous batch update
func BatchUpdateExpenses(c fiber.Ctx) error {
	log.Println("Batch Update expense triggered")
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	var updates []mdlFeatureOne.BatchUpdateItem
	if err := c.Bind().Body(&updates); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate batch size
	if len(updates) == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"No updates provided", nil, http.StatusBadRequest)
	}
	if len(updates) > 100 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Batch size too large. Maximum 100 updates allowed", nil, http.StatusBadRequest)
	}

	// Process updates
	var results []mdlFeatureOne.BatchUpdateResultItem
	successCount := 0
	failCount := 0

	for i, update := range updates {
		// Validate expense ID
		if update.ExpenseID == 0 {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:     i,
				ExpenseID: update.ExpenseID,
				Message:   "Expense ID is required",
				Success:   false,
			})
			continue
		}

		if !scpFeatureOne.ExpenseExists(userID, update.ExpenseID) {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:     i,
				ExpenseID: update.ExpenseID,
				Message:   "Expense not found",
				Success:   false,
			})
			continue
		}

		// Build update request
		req := &mdlFeatureOne.UpdateExpenseRequest{
			Title:      update.Title,
			Amount:     update.Amount,
			CategoryID: update.CategoryID,
			Date:       update.Date,
			Notes:      update.Notes,
		}

		// Attempt update
		_, err := scpFeatureOne.UpdateExpense(userID, update.ExpenseID, req)
		if err != nil {
			failCount++
			results = append(results, mdlFeatureOne.BatchUpdateResultItem{
				Index:     i,
				ExpenseID: update.ExpenseID,
				Message:   "Failed to update expense",
				Success:   false,
			})
		} else {
			successCount++
		}
	}

	// Build response
	response := mdlFeatureOne.BatchUpdateResponse{
		Results:    results,
		Total:      len(updates),
		Successful: successCount,
		Failed:     failCount,
	}

	if failCount > 0 {
		return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
			"Batch update completed with some errors", response, http.StatusMultiStatus)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"All expenses updated successfully", response, http.StatusOK)
}

// BatchUpdateExpensesAsync performs asynchronous batch update
func BatchUpdateExpensesAsync(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	var updates []mdlFeatureOne.BatchUpdateItem
	if err := c.Bind().Body(&updates); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate batch size
	if len(updates) == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"No updates provided", nil, http.StatusBadRequest)
	}
	if len(updates) > 100 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Batch size too large. Maximum 100 updates allowed", nil, http.StatusBadRequest)
	}

	// Create batch job
	jobID, err := scpFeatureOne.CreateBatchJob(userID, "expense_batch_update", len(updates))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create batch job", err, http.StatusInternalServerError)
	}

	// Process in background
	go scpFeatureOne.ProcessBatchUpdate(jobID, userID, updates)

	// Return job info immediately
	response := mdlFeatureOne.BatchJobCreatedResponse{
		JobID:      jobID,
		TotalItems: len(updates),
		Status:     "pending",
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_202,
		"Batch update job created successfully", response, http.StatusAccepted)
}

// ============================================
// BATCH UPLOAD ENDPOINTS
// ============================================

// BatchUploadExpensesFromCSV uploads expenses from CSV file
func BatchUploadExpensesFromCSV(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	// Parse uploaded CSV file
	file, err := c.FormFile("file")
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"CSV file is required", err, http.StatusBadRequest)
	}

	f, err := file.Open()
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to open CSV file", err, http.StatusInternalServerError)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid CSV format", err, http.StatusBadRequest)
	}

	if len(records) < 2 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"CSV must contain header and at least one row", nil, http.StatusBadRequest)
	}

	// Validate CSV headers
	expectedHeaders := []string{"title", "amount", "categoryid", "date", "notes"}
	headers := records[0]

	if len(headers) != len(expectedHeaders) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			fmt.Sprintf("Invalid CSV header count. Expected headers: %v", expectedHeaders),
			nil, http.StatusBadRequest)
	}

	for i, expected := range expectedHeaders {
		if strings.ToLower(strings.TrimSpace(headers[i])) != expected {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				fmt.Sprintf("Invalid CSV header at column %d: expected '%s', got '%s'",
					i+1, expected, headers[i]),
				nil, http.StatusBadRequest)
		}
	}

	// Parse CSV data
	expenses := make([]mdlFeatureOne.CSVExpenseRow, 0, len(records)-1)
	for i, row := range records[1:] {
		if len(row) != len(headers) {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				fmt.Sprintf("Row %d has mismatched columns", i+2), nil, http.StatusBadRequest)
		}

		// Parse amount
		amount, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				fmt.Sprintf("Invalid amount at row %d", i+2), err, http.StatusBadRequest)
		}

		// Parse category ID (optional)
		var categoryID *int
		if strings.TrimSpace(row[2]) != "" {
			catID, err := strconv.Atoi(strings.TrimSpace(row[2]))
			if err != nil {
				return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
					fmt.Sprintf("Invalid category ID at row %d", i+2), err, http.StatusBadRequest)
			}
			categoryID = &catID
		}

		// Parse notes (optional)
		var notes *string
		if strings.TrimSpace(row[4]) != "" {
			n := strings.TrimSpace(row[4])
			notes = &n
		}

		expense := mdlFeatureOne.CSVExpenseRow{
			Title:      strings.TrimSpace(row[0]),
			Amount:     amount,
			CategoryID: categoryID,
			Date:       strings.TrimSpace(row[3]),
			Notes:      notes,
		}
		expenses = append(expenses, expense)
	}

	if len(expenses) > 1000 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Batch too large (max 1000 rows)", nil, http.StatusBadRequest)
	}

	// Create batch job
	jobID, err := scpFeatureOne.CreateBatchJob(userID, "expense_batch_upload_csv", len(expenses))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create batch job", err, http.StatusInternalServerError)
	}

	// Process in background
	go scpFeatureOne.ProcessBatchUpload(jobID, userID, expenses)

	// Return job info immediately
	response := mdlFeatureOne.BatchJobCreatedResponse{
		JobID:      jobID,
		TotalItems: len(expenses),
		Status:     "pending",
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_202,
		"CSV batch upload job created successfully", response, http.StatusAccepted)
}

// ============================================
// BATCH JOB STATUS ENDPOINTS
// ============================================

// GetBatchJobStatus retrieves batch job status
func GetBatchJobStatus(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	jobID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid job ID", err, http.StatusBadRequest)
	}

	// Check if job exists
	if !scpFeatureOne.BatchJobExists(userID, jobID) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_404,
			"Job not found", nil, http.StatusNotFound)
	}

	// Get job status
	job, err := scpFeatureOne.GetBatchJob(userID, jobID)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to retrieve job status", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Job status retrieved successfully", job, http.StatusOK)
}
