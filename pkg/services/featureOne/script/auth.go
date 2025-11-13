package scpFeatureOne

import (
	"encoding/json"
	"fmt"
	"go_template_v3/pkg/config"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	"log"
	"time"
)

// ============================================
// USER OPERATIONS
// ============================================

// UserExists checks if a user exists by ID
func UserExists(userID int) bool {
	var exists bool

	err := config.DBConnList[0].Raw(
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		userID,
	).Scan(&exists).Error

	if err != nil {
		log.Printf("[UserExists] Error checking user %d: %v", userID, err)
		return false
	}

	return exists
}

// UserExistsByEmail checks if a user exists by email
func UserExistsByEmail(email string) bool {
	var exists bool

	err := config.DBConnList[0].Raw(
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`,
		email,
	).Scan(&exists).Error

	if err != nil {
		log.Printf("[UserExistsByEmail] Error checking email %s: %v", email, err)
		return false
	}

	return exists
}

// RegisterUser creates a new user in the database
func RegisterUser(req *mdlFeatureOne.RegisterRequest) (*mdlFeatureOne.UserResponse, error) {
	var jsonResult string
	var user mdlFeatureOne.UserResponse

	// PostgreSQL function returns JSONB
	err := config.DBConnList[0].Raw(
		`SELECT register_user($1, $2, $3)::text`,
		req.Email,
		req.Password,
		req.Name,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[RegisterUser] Error calling function: %v", err)
		return nil, err
	}
	// Decode JSONB to struct
	if err := json.Unmarshal([]byte(jsonResult), &user); err != nil {
		log.Printf("[RegisterUser] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[RegisterUser] Success - UserID: %d, Email: %s", user.ID, user.Email)
	return &user, nil
}

// GetUserByEmail retrieves user by email
func GetUserByEmail(email string) (*mdlFeatureOne.UserEntity, error) {
	var jsonResult string
	var user mdlFeatureOne.UserEntity

	err := config.DBConnList[0].Raw(
		`SELECT get_user_by_email($1)`,
		email,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[GetUserByEmail] Error for email %s: %v", email, err)
		return nil, err
	}

	// Decode JSONB to struct
	if err := json.Unmarshal([]byte(jsonResult), &user); err != nil {
		log.Printf("[GetUserByEmail] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[GetUserByEmail] Found user - ID: %d, Email: %s", user.ID, user.Email)
	return &user, nil
}

// GetUserByID retrieves user by ID
func GetUserByID(userID int) (*mdlFeatureOne.UserEntity, error) {
	var user mdlFeatureOne.UserEntity

	err := config.DBConnList[0].Raw(
		`SELECT id, email, password, name, created_at, updated_at FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&user).Error

	if err != nil {
		log.Printf("[GetUserByID] Error for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("[GetUserByID] Found user - ID: %d, Email: %s", user.ID, user.Email)
	return &user, nil
}

// UpdateUser updates user information
func UpdateUser(userID int, req *mdlFeatureOne.UpdateUserRequest, hashedPassword *string) (*mdlFeatureOne.UserEntity, error) {
	var user mdlFeatureOne.UserEntity
	var jsonResult string

	err := config.DBConnList[0].Raw(
		`SELECT * FROM update_user($1, $2, $3)`,
		userID,
		req.Name,
		hashedPassword,
	).Scan(&jsonResult).Error

	if err != nil {
		log.Printf("[UpdateUser] Error updating user %d: %v", userID, err)
		return nil, err
	}

	// Decode JSONB to struct
	if err := json.Unmarshal([]byte(jsonResult), &user); err != nil {
		log.Printf("[UpdateUser] JSON parse error: %v", err)
		return nil, err
	}

	log.Printf("[UpdateUser] Success - UserID: %d, New Name: %s", user.ID, user.Name)
	return &user, nil
}

// ============================================
// PASSWORD RESET OPERATIONS
// ============================================

// CreateResetToken creates a password reset token
func CreateResetToken(userID int, tokenHash string, expiresAt time.Time) (int, error) {
	db := config.DBConnList[0]

	// Step 1: Invalidate old unused tokens
	err := db.Exec(`
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND used_at IS NULL
	`, userID).Error
	if err != nil {
		log.Printf("[CreateResetToken] Error invalidating old tokens for user %d: %v", userID, err)
		return 0, err
	}

	// Step 2: Insert new token and return its ID
	var tokenID int
	err = db.Raw(`
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
		RETURNING id
	`, userID, tokenHash, expiresAt).Scan(&tokenID).Error
	if err != nil {
		log.Printf("[CreateResetToken] Error creating token for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("[CreateResetToken] Success - TokenID: %d, UserID: %d", tokenID, userID)
	return tokenID, nil
}

// VerifyResetToken verifies if token is valid and not expired
func VerifyResetToken(tokenHash string) (*mdlFeatureOne.ResetTokenVerification, error) {
	var verification mdlFeatureOne.ResetTokenVerification

	// Query the valid reset token
	err := config.DBConnList[0].Raw(`
		SELECT prt.id AS token_id, prt.user_id AS user_id
		FROM password_reset_tokens prt
		JOIN users u ON prt.user_id = u.id
		WHERE prt.token_hash = ?
		  AND prt.used_at IS NULL
		  AND prt.expires_at > CURRENT_TIMESTAMP
		  AND u.deleted_at IS NULL
		LIMIT 1
	`, tokenHash).Scan(&verification).Error

	if err != nil {
		log.Printf("[VerifyResetToken] Error verifying token: %v", err)
		return nil, err
	}

	// Check if no valid token found
	if verification.TokenID == 0 {
		return nil, fmt.Errorf("invalid or expired token")
	}

	log.Printf("[VerifyResetToken] Valid token - TokenID: %d, UserID: %d", verification.TokenID, verification.UserID)
	return &verification, nil
}

// ResetPassword updates user password and marks token as used
func ResetPassword(tokenID, userID int, newPassword string) error {
	db := config.DBConnList[0]

	// Step 1: Update user's password
	err := db.Exec(`
		UPDATE users
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
	`, newPassword, userID).Error
	if err != nil {
		log.Printf("[ResetPassword] Error updating password for user %d: %v", userID, err)
		return err
	}

	// Step 2: Mark the reset token as used
	err = db.Exec(`
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, tokenID).Error
	if err != nil {
		log.Printf("[ResetPassword] Error marking token %d as used: %v", tokenID, err)
		return err
	}

	log.Printf("[ResetPassword] Success - UserID: %d, TokenID: %d", userID, tokenID)
	return nil
}
