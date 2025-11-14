package ctrFeatureOne

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	v1 "github.com/FDSAP-Git-Org/hephaestus/helper/v1"
	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	utils_v1 "github.com/FDSAP-Git-Org/hephaestus/utils/v1"
	"github.com/gofiber/fiber/v3"

	"go_template_v3/pkg/config"
	"go_template_v3/pkg/global/utils"
	hlpFeatureOne "go_template_v3/pkg/services/featureOne/helper"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	scpFeatureOne "go_template_v3/pkg/services/featureOne/script"
)

// ============================================
// EXPENSE ENDPOINTS
// ============================================

// CreateExpense creates a new expense
func CreateExpense(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	var req mdlFeatureOne.CreateExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate
	if err := hlpFeatureOne.ValidateCreateExpense(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			err.Error(), nil, http.StatusBadRequest)
	}

	// Create expense
	expense, err := scpFeatureOne.CreateExpense(userID, &req)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create expense", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_201,
		"Expense created successfully", expense, http.StatusCreated)
}

// CreateExpenseV2 creates expense with file upload support
func CreateExpenseV2(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid form data", err, http.StatusBadRequest)
	}

	// Bind form data
	var req mdlFeatureOne.CreateExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate
	if err := hlpFeatureOne.ValidateCreateExpense(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			err.Error(), nil, http.StatusBadRequest)
	}

	// Handle file upload
	if files, ok := form.File["image"]; ok && len(files) > 0 {
		fileHeader := files[0]
		config := utils.DefaultFileUploadConfig()

		uploadedPath, err := utils.UploadFile(c, fileHeader, config)
		if err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Failed to upload image", err, http.StatusBadRequest)
		}

		fullURL := utils_v1.GetEnv("BASE_URL") + uploadedPath
		req.ImageURL = &fullURL
	}

	// Create expense
	expense, err := scpFeatureOne.CreateExpense(userID, &req)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create expense", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_201,
		"Expense created successfully", expense, http.StatusCreated)
}

func CreateExpenseWithCloudinary(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid form data", err, http.StatusBadRequest)
	}

	// Bind form data
	var req mdlFeatureOne.CreateExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate
	if err := hlpFeatureOne.ValidateCreateExpense(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			err.Error(), nil, http.StatusBadRequest)
	}

	// Handle file upload
	if files, ok := form.File["image"]; ok && len(files) > 0 {
		fileHeader := files[0]
		cnf := config.LoadCloudinaryConfig()

		uploadedPath, err := config.UploadToCloudinary(fileHeader, cnf)
		if err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Failed to upload image", err, http.StatusBadRequest)
		}

		req.ImageURL = &uploadedPath
	}

	// Create expense
	expense, err := scpFeatureOne.CreateExpense(userID, &req)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create expense", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_201,
		"Expense created successfully", expense, http.StatusCreated)
}

// GetExpenses retrieves expenses with filters
func GetExpenses(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	// Parse query parameters
	filters := &mdlFeatureOne.ExpenseFilters{
		Title:      getQueryString(c, "title"),
		MinAmount:  getQueryFloat(c, "minAmount"),
		MaxAmount:  getQueryFloat(c, "maxAmount"),
		CategoryID: getQueryInt(c, "categoryId"),
		StartDate:  getQueryString(c, "startDate"),
		EndDate:    getQueryString(c, "endDate"),
		Limit:      getQueryIntDefault(c, "limit", 50),
		Offset:     getQueryIntDefault(c, "offset", 0),
	}

	// Validate limit
	if filters.Limit < 1 || filters.Limit > 100 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Limit must be between 1 and 100", nil, http.StatusBadRequest)
	}

	// Get expenses
	result, err := scpFeatureOne.GetExpenses(userID, filters)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to retrieve expenses", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Expenses retrieved successfully", result, http.StatusOK)
}

// GetExpense retrieves a single expense by ID
func GetExpense(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	expenseID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid expense ID", err, http.StatusBadRequest)
	}

	// Check if expense exists
	if !scpFeatureOne.ExpenseExists(userID, expenseID) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_404,
			"Expense not found", nil, http.StatusNotFound)
	}

	// Get expense
	expense, err := scpFeatureOne.GetExpenseByID(userID, expenseID)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to retrieve expense", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Expense retrieved successfully", expense, http.StatusOK)
}

// UpdateExpense updates an existing expense
func UpdateExpense(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	expenseID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid expense ID", err, http.StatusBadRequest)
	}

	var req mdlFeatureOne.UpdateExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate at least one field provided
	if req.Title == nil && req.Amount == nil && req.CategoryID == nil &&
		req.Date == nil && req.Notes == nil && req.ImageURL == nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"At least one field to update is required", nil, http.StatusBadRequest)
	}

	// Validate amount if provided
	if req.Amount != nil && *req.Amount <= 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Amount must be greater than 0", nil, http.StatusBadRequest)
	}

	// Validate date if provided
	if req.Date != nil && strings.TrimSpace(*req.Date) != "" {
		if _, err := time.Parse("2006-01-02", *req.Date); err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Invalid date format (expected YYYY-MM-DD)", err, http.StatusBadRequest)
		}
	}

	// Check if expense exists
	if !scpFeatureOne.ExpenseExists(userID, expenseID) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_404,
			"Expense not found", nil, http.StatusNotFound)
	}

	// Update expense
	expense, err := scpFeatureOne.UpdateExpense(userID, expenseID, &req)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to update expense", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Expense updated successfully", expense, http.StatusOK)
}

// DeleteExpense soft deletes an expense
func DeleteExpense(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	expenseID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid expense ID", err, http.StatusBadRequest)
	}

	// Check if expense exists
	if !scpFeatureOne.ExpenseExists(userID, expenseID) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_404,
			"Expense not found", nil, http.StatusNotFound)
	}

	// Delete expense
	result, err := scpFeatureOne.DeleteExpense(userID, expenseID)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to delete expense", err, http.StatusInternalServerError)
	}

	if !result.IsDeleted {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to delete expense", nil, http.StatusInternalServerError)
	}

	// Delete image file if exists
	if result.ImageURL != nil && *result.ImageURL != "" {
		if err := utils.DeleteUploadedFile(*result.ImageURL); err != nil {
			fmt.Printf("Warning: Failed to delete image file %s: %v\n", *result.ImageURL, err)
		}
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Expense deleted successfully", nil, http.StatusOK)
}

func DeleteExpenseWithCloudinary(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	expenseID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid expense ID", err, http.StatusBadRequest)
	}

	if !scpFeatureOne.ExpenseExists(userID, expenseID) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_404,
			"Expense not found", nil, http.StatusNotFound)
	}

	result, err := scpFeatureOne.DeleteExpense(userID, expenseID)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to delete expense", err, http.StatusInternalServerError)
	}

	if !result.IsDeleted {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to delete expense", nil, http.StatusInternalServerError)
	}

	// Delete Cloudinary image if exists
	if result.ImageURL != nil && *result.ImageURL != "" {
		publicID := config.CloudinaryPublicIDFromURL(*result.ImageURL)
		if publicID != "" {
			cloudCfg := config.LoadCloudinaryConfig()
			if err := config.DeleteCloudinaryImage(publicID, cloudCfg); err != nil {
				fmt.Printf("Warning: Failed to delete Cloudinary image %s: %v\n", *result.ImageURL, err)
			}
		}
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Expense deleted successfully", nil, http.StatusOK)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// getQueryString gets string query parameter
func getQueryString(c fiber.Ctx, key string) *string {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	return &val
}

// getQueryInt gets int query parameter
func getQueryInt(c fiber.Ctx, key string) *int {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &num
}

// getQueryFloat gets float query parameter
func getQueryFloat(c fiber.Ctx, key string) *float64 {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil
	}
	return &num
}

// getQueryIntDefault gets int with default value
func getQueryIntDefault(c fiber.Ctx, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return num
}
