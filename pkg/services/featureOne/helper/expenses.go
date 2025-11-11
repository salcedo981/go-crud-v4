package hlpFeatureOne

import (
	"fmt"
	mdlFeatureOne "go_template_v3/pkg/services/featureOne/model"
	"strings"
	"time"
)

func ValidateCreateExpense(req *mdlFeatureOne.CreateExpenseRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if strings.TrimSpace(req.Date) == "" {
		return fmt.Errorf("date is required")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD)")
	}
	return nil
}
