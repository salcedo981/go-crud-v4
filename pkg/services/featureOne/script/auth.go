package scpFeatureOne

import (
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
func RegisterUser(req *mdlFeatureOne.RegisterRequest) (*mdlFeatureOne.UserEntity, error) {
	var user mdlFeatureOne.UserEntity

	err := config.DBConnList[0].Raw(
		`SELECT * FROM register_user($1, $2, $3)`,
		req.Email,
		req.Password,
		req.Name,
	).Scan(&user).Error

	if err != nil {
		log.Printf("[RegisterUser] Error: %v", err)
		return nil, err
	}

	log.Printf("[RegisterUser] Success - UserID: %d, Email: %s", user.ID, user.Email)
	return &user, nil
}

// GetUserByEmail retrieves user by email
func GetUserByEmail(email string) (*mdlFeatureOne.UserEntity, error) {
	var user mdlFeatureOne.UserEntity

	err := config.DBConnList[0].Raw(
		`SELECT * FROM get_user_by_email($1)`,
		email,
	).Scan(&user).Error

	if err != nil {
		log.Printf("[GetUserByEmail] Error for email %s: %v", email, err)
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

	err := config.DBConnList[0].Raw(
		`SELECT * FROM update_user($1, $2, $3)`,
		userID,
		req.Name,
		hashedPassword,
	).Scan(&user).Error

	if err != nil {
		log.Printf("[UpdateUser] Error updating user %d: %v", userID, err)
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
	var tokenID int

	err := config.DBConnList[0].Raw(
		`SELECT create_reset_token($1, $2, $3)`,
		userID,
		tokenHash,
		expiresAt,
	).Scan(&tokenID).Error

	if err != nil {
		log.Printf("[CreateResetToken] Error for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("[CreateResetToken] Success - TokenID: %d, UserID: %d", tokenID, userID)
	return tokenID, nil
}

// VerifyResetToken verifies if token is valid and not expired
func VerifyResetToken(tokenHash string) (*mdlFeatureOne.ResetTokenVerification, error) {
	var verification mdlFeatureOne.ResetTokenVerification

	err := config.DBConnList[0].Raw(
		`SELECT * FROM verify_reset_token($1)`,
		tokenHash,
	).Scan(&verification).Error

	if err != nil {
		log.Printf("[VerifyResetToken] Token verification failed: %v", err)
		return nil, err
	}

	if verification.TokenID == 0 {
		return nil, fmt.Errorf("invalid or expired token")
	}

	log.Printf("[VerifyResetToken] Valid token - TokenID: %d, UserID: %d",
		verification.TokenID, verification.UserID)
	return &verification, nil
}

// ResetPassword updates user password and marks token as used
func ResetPassword(tokenID, userID int, newPassword string) error {
	var success bool

	err := config.DBConnList[0].Raw(
		`SELECT reset_password($1, $2, $3)`,
		tokenID,
		userID,
		newPassword,
	).Scan(&success).Error

	if err != nil {
		log.Printf("[ResetPassword] Error for user %d: %v", userID, err)
		return err
	}

	if !success {
		log.Printf("[ResetPassword] Failed for user %d", userID)
		return err
	}

	log.Printf("[ResetPassword] Success - UserID: %d", userID)
	return nil
}
