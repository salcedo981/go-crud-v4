package ctrFeatureOne

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	v1 "github.com/FDSAP-Git-Org/hephaestus/helper/v1"
	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	utils_v1 "github.com/FDSAP-Git-Org/hephaestus/utils/v1"
	"github.com/gofiber/fiber/v3"

	"go_template_v3/pkg/global/utils"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	scpFeatureOne "go_template_v3/pkg/services/featureOne/script"
)

// ============================================
// AUTH ENDPOINTS
// ============================================

// Register creates a new user account
func Register(c fiber.Ctx) error {
	var req mdlFeatureOne.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Email is required", nil, http.StatusBadRequest)
	}
	if !utils_v1.IsEmailValid(req.Email) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid email format", nil, http.StatusBadRequest)
	}

	// Check if user already exists
	if scpFeatureOne.UserExistsByEmail(req.Email) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Email already exists", nil, http.StatusBadRequest)
	}

	// Validate password
	if strings.TrimSpace(req.Password) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Password is required", nil, http.StatusBadRequest)
	}
	if !utils_v1.IsPasswordValid(req.Password) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Password must be 8+ chars with uppercase, lowercase, and special char",
			nil, http.StatusBadRequest)
	}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Name is required", nil, http.StatusBadRequest)
	}

	// Hash password
	hashedPassword, err := utils_v1.HashData(req.Password)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to hash password", err, http.StatusInternalServerError)
	}
	req.Password = hashedPassword

	// Register user
	user, err := scpFeatureOne.RegisterUser(&req)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Registration failed", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_201,
		"User registered successfully", user, http.StatusCreated)
}

// Login authenticates user and returns JWT token
func Login(c fiber.Ctx) error {
	var req mdlFeatureOne.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate request
	if strings.TrimSpace(req.Email) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Email is required", nil, http.StatusBadRequest)
	}
	if strings.TrimSpace(req.Password) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Password is required", nil, http.StatusBadRequest)
	}

	// Check if user exists
	if !scpFeatureOne.UserExistsByEmail(req.Email) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Invalid credentials", nil, http.StatusUnauthorized)
	}

	// Get user by email
	user, err := scpFeatureOne.GetUserByEmail(req.Email)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to retrieve user", err, http.StatusInternalServerError)
	}

	// Verify password
	if !utils_v1.CheckHashData(req.Password, user.Password) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Invalid credentials", nil, http.StatusUnauthorized)
	}

	// Generate JWT token
	claims := map[string]interface{}{
		"userId": user.ID,
		"email":  user.Email,
		"name":   user.Name,
	}

	token, err := utils_v1.GenerateJWTSignedString(
		[]byte(utils_v1.GetEnv("SECRET_KEY")),
		24, // 24 hours
		claims,
	)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Token generation failed", err, http.StatusInternalServerError)
	}

	// Map to response
	response := mdlFeatureOne.LoginResponse{
		Token: token,
		User: mdlFeatureOne.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Login successful", response, http.StatusOK)
}

// Logout handles user logout (JWT is stateless, so this is optional)
func Logout(c fiber.Ctx) error {
	log.Printf("Logout hit")
	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Logout successful", nil, http.StatusOK)
}

// UpdateUser updates user profile information and optionally password
func UpdateUser(c fiber.Ctx) error {
	userID := utils.GetUserId(c)
	if userID == 0 {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_401,
			"Unauthorized", nil, http.StatusUnauthorized)
	}

	var req mdlFeatureOne.UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Name is required", nil, http.StatusBadRequest)
	}

	// If password change is requested, validate it
	var hashedPassword *string
	if req.OldPassword != nil || req.NewPassword != nil {
		// Both old and new password must be provided
		if req.OldPassword == nil || req.NewPassword == nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Both oldPassword and newPassword are required to change password",
				nil, http.StatusBadRequest)
		}

		// Validate old password is not empty
		if strings.TrimSpace(*req.OldPassword) == "" {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Old password cannot be empty", nil, http.StatusBadRequest)
		}

		// Validate new password
		if strings.TrimSpace(*req.NewPassword) == "" {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"New password cannot be empty", nil, http.StatusBadRequest)
		}
		if !utils_v1.IsPasswordValid(*req.NewPassword) {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"New password must be 8+ chars with uppercase, lowercase, and special char",
				nil, http.StatusBadRequest)
		}

		// Get current user to verify old password
		currentUser, err := scpFeatureOne.GetUserByID(userID)
		if err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
				"Failed to verify user", err, http.StatusInternalServerError)
		}

		// Verify old password
		if !utils_v1.CheckHashData(*req.OldPassword, currentUser.Password) {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
				"Incorrect old password", nil, http.StatusBadRequest)
		}

		// Hash new password
		hashed, err := utils_v1.HashData(*req.NewPassword)
		if err != nil {
			return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
				"Failed to hash password", err, http.StatusInternalServerError)
		}
		hashedPassword = &hashed
	}

	// Update user
	user, err := scpFeatureOne.UpdateUser(userID, &req, hashedPassword)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to update user", err, http.StatusInternalServerError)
	}

	// Map to response
	response := mdlFeatureOne.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	message := "User updated successfully"
	if hashedPassword != nil {
		message = "User updated successfully (including password)"
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		message, response, http.StatusOK)
}

// ============================================
// PASSWORD RESET ENDPOINTS
// ============================================

// ForgotPassword initiates password reset process
func ForgotPassword(c fiber.Ctx) error {
	var req mdlFeatureOne.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate email
	if strings.TrimSpace(req.Email) == "" || !utils_v1.IsEmailValid(req.Email) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Valid email required", nil, http.StatusBadRequest)
	}

	// Check if user exists (but don't reveal this to user for security)
	if !scpFeatureOne.UserExistsByEmail(req.Email) {
		// Return success even if user not found (security best practice)
		return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
			"If email exists, reset link sent", nil, http.StatusOK)
	}

	// Get user
	user, err := scpFeatureOne.GetUserByEmail(req.Email)
	if err != nil {
		// Return success even on error (security best practice)
		return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
			"If email exists, reset link sent", nil, http.StatusOK)
	}

	// Generate reset token
	token := utils_v1.GenerateRandomStrings(32, []string{
		utils_v1.UpperString,
		utils_v1.LowerString,
		utils_v1.NumericString,
	})
	tokenHash := utils_v1.HashDataSHA512(token)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create reset token
	_, err = scpFeatureOne.CreateResetToken(user.ID, tokenHash, expiresAt)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to create reset token", err, http.StatusInternalServerError)
	}

	// Send email with reset link (async)
	go func() {
		if err := sendPasswordResetEmail(user.Email, user.Name, token); err != nil {
			// Log the error but don't fail the request for security reasons
			fmt.Printf("Failed to send reset email to %s: %v\n", user.Email, err)
		}
	}()

	// In development, return token for testing
	response := map[string]interface{}{
		"token": token,
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Password reset initiated", response, http.StatusOK)
}

// VerifyResetToken verifies if reset token is valid
func VerifyResetToken(c fiber.Ctx) error {
	var req mdlFeatureOne.VerifyResetTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	if strings.TrimSpace(req.Token) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Token is required", nil, http.StatusBadRequest)
	}

	// Verify token
	tokenHash := utils_v1.HashDataSHA512(req.Token)
	_, err := scpFeatureOne.VerifyResetToken(tokenHash)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid or expired token", err, http.StatusBadRequest)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Token is valid", nil, http.StatusOK)
}

// ResetPassword resets user password with valid token
func ResetPassword(c fiber.Ctx) error {
	var req mdlFeatureOne.ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validate request
	if strings.TrimSpace(req.Token) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Token is required", nil, http.StatusBadRequest)
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"New password is required", nil, http.StatusBadRequest)
	}
	if !utils_v1.IsPasswordValid(req.NewPassword) {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Password does not meet requirements", nil, http.StatusBadRequest)
	}

	// Verify token
	tokenHash := utils_v1.HashDataSHA512(req.Token)
	verification, err := scpFeatureOne.VerifyResetToken(tokenHash)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_400,
			"Invalid or expired token", err, http.StatusBadRequest)
	}

	// Hash new password
	hashedPassword, err := utils_v1.HashData(req.NewPassword)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to hash password", err, http.StatusInternalServerError)
	}

	// Reset password
	err = scpFeatureOne.ResetPassword(verification.TokenID, verification.UserID, hashedPassword)
	if err != nil {
		return v1.JSONResponseWithError(c, respcode.ERR_CODE_500,
			"Failed to reset password", err, http.StatusInternalServerError)
	}

	return v1.JSONResponseWithData(c, respcode.SUC_CODE_200,
		"Password reset successfully", nil, http.StatusOK)
}

// ============================================
// SEND MAIL HELPER FUNCTIONS
// ============================================

func sendPasswordResetEmail(email, name, token string) error {
	// Get frontend URL from environment variables
	frontendURL := utils_v1.GetEnv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // default for development
	}

	// Create reset link with token as query parameter
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	// Email content
	subject := "Password Reset Request"

	// HTML email template
	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.button { display: inline-block; padding: 12px 24px; background-color: #007bff; 
					color: white !important; text-decoration: none; border-radius: 4px; margin: 20px 0; }
			.footer { margin-top: 30px; font-size: 12px; color: #666; }
		</style>
	</head>
	<body>
		<div class="container">
			<h2>Password Reset Request</h2>
			<p>Hello %s,</p>
			<p>You requested to reset your password. Click the button below to create a new password:</p>
			<p><a href="%s" class="button">Reset Password</a></p>
			<p>Or copy and paste this link in your browser:</p>
			<p><code>%s</code></p>
			<p>This link will expire in 1 hour for security reasons.</p>
			<p>If you didn't request this reset, please ignore this email.</p>
			<div class="footer">
				<p>This is an automated message, please do not reply to this email.</p>
			</div>
		</div>
	</body>
	</html>
	`, name, resetLink, resetLink)

	fmt.Printf("=== PASSWORD RESET EMAIL ===\n")
	fmt.Printf("To: %s\n", email)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Reset Link: %s\n", resetLink)
	fmt.Printf("=======================\n")

	// Send via SMTP
	return sendWithSMTP(email, subject, htmlBody)
}

func sendWithSMTP(to, subject, htmlContent string) error {
	smtpHost := utils_v1.GetEnv("SMTP_HOST")
	smtpPort := utils_v1.GetEnv("SMTP_PORT")
	smtpUser := utils_v1.GetEnv("SMTP_USER")
	smtpPass := utils_v1.GetEnv("SMTP_PASS")
	from := utils_v1.GetEnv("EMAIL_FROM")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + htmlContent)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	return err
}
