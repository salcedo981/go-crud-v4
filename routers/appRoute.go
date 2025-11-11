package routers

import (
	"go_template_v3/pkg/middleware"
	// ctrEncryption "go_template_v3/pkg/services/encryption/controller"
	ctrFeatureOne "go_template_v3/pkg/services/featureOne/controller"
	svcHealthcheck "go_template_v3/pkg/services/healthcheck"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// APIRoute configures all application routes
func APIRoute(app *fiber.App) {
	// Serve static files (for uploaded images)
	app.Use("/assets", static.New("./assets", static.Config{
		MaxAge: 3600, // 1 hour cache
	}))

	// API route groups
	publicV1 := app.Group("/api/public/v1")
	privateV1 := app.Group("/api/private/v1")

	// ============================================
	// HEALTH CHECK
	// ============================================
	publicV1.Get("/", svcHealthcheck.HealthCheck)
	privateV1.Get("/", svcHealthcheck.HealthCheck)

	// ============================================
	// AUTHENTICATION ROUTES (PUBLIC)
	// ============================================
	authGroup := publicV1.Group("/auth")
	authGroup.Post("/register", ctrFeatureOne.Register)
	authGroup.Post("/login", ctrFeatureOne.Login)
	authGroup.Post("/forgot-password", ctrFeatureOne.ForgotPassword)
	authGroup.Post("/verify-reset-token", ctrFeatureOne.VerifyResetToken)
	authGroup.Post("/reset-password", ctrFeatureOne.ResetPassword)

	// ============================================
	// AUTHENTICATION ROUTES (PROTECTED)
	// ============================================
	authProtected := publicV1.Group("/auth", middleware.AuthMiddleware)
	authProtected.Put("/update-user", ctrFeatureOne.UpdateUser)
	authProtected.Post("/logout", ctrFeatureOne.Logout)

	// ============================================
	// CATEGORY ROUTES (PUBLIC)
	// ============================================
	categoryGroup := publicV1.Group("/expense-categories")
	categoryGroup.Post("/", ctrFeatureOne.CreateCategory)
	categoryGroup.Get("/", ctrFeatureOne.GetCategories)

	// ============================================
	// EXPENSE ROUTES (PROTECTED)
	// ============================================
	expenseGroup := publicV1.Group("/expenses", middleware.AuthMiddleware)

	// Batch operations
	expenseGroup.Put("/batch", ctrFeatureOne.BatchUpdateExpenses)            // Sync
	expenseGroup.Put("/batch-async", ctrFeatureOne.BatchUpdateExpensesAsync) // Async
	expenseGroup.Get("/batch-async/:id", ctrFeatureOne.GetBatchJobStatus)
	expenseGroup.Post("/batch-upload", ctrFeatureOne.BatchUploadExpensesFromCSV) // CSV Upload

	// Basic CRUD
	expenseGroup.Post("/", ctrFeatureOne.CreateExpense)
	expenseGroup.Post("/v2", ctrFeatureOne.CreateExpenseV2) // With file upload
	expenseGroup.Get("/", ctrFeatureOne.GetExpenses)
	expenseGroup.Get("/:id", ctrFeatureOne.GetExpense)
	expenseGroup.Put("/:id", ctrFeatureOne.UpdateExpense)
	expenseGroup.Delete("/:id", ctrFeatureOne.DeleteExpense)

	// ============================================
	// BATCH JOB ROUTES (PROTECTED)
	// ============================================
	batchGroup := publicV1.Group("/batch", middleware.AuthMiddleware)
	batchGroup.Get("/jobs/:jobId", ctrFeatureOne.GetBatchJobStatus)
}
