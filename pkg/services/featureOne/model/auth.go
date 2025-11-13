package mdlFeatureOne

// ============================================
// AUTH REQUEST STRUCTS
// ============================================

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name        string  `json:"name"`
	OldPassword *string `json:"oldPassword"`
	NewPassword *string `json:"newPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type VerifyResetTokenRequest struct {
	Token string `json:"token"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// ============================================
// AUTH RESPONSE STRUCTS
// ============================================

type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

// ============================================
// AUTH ENTITY STRUCTS (DB)
// ============================================

type UserEntity struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type PasswordResetTokenEntity struct {
	ID        int     `db:"id"`
	UserID    int     `db:"user_id"`
	TokenHash string  `db:"token_hash"`
	ExpiresAt string  `db:"expires_at"`
	UsedAt    *string `db:"used_at"`
	CreatedAt string  `db:"created_at"`
}

// ============================================
// HELPER STRUCTS
// ============================================

type ResetTokenVerification struct {
	TokenID int `db:"token_id"`
	UserID  int `db:"user_id"`
}
