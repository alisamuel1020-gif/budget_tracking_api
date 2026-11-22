package main

import (
	"errors"
	"strings"
	"time"
)

const (
	TypeIncome  = "income"
	TypeExpense = "expense"
)

// Transaction is one income or expense entry.
// Amounts are stored in cents so money never suffers floating-point rounding errors.
type Transaction struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amountCents"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Date        string `json:"date"` // YYYY-MM-DD
}

// TransactionInput is the body a client sends to create a transaction.
type TransactionInput struct {
	Description string `json:"description"`
	AmountCents int64  `json:"amountCents"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Date        string `json:"date"`
}

// Validate returns an error describing the first invalid field, or nil.
func (in TransactionInput) Validate() error {
	if strings.TrimSpace(in.Description) == "" {
		return errors.New("description is required")
	}
	if len(in.Description) > 200 {
		return errors.New("description must be 200 characters or fewer")
	}
	if in.AmountCents <= 0 {
		return errors.New("amountCents must be greater than 0")
	}
	if in.Type != TypeIncome && in.Type != TypeExpense {
		return errors.New(`type must be "income" or "expense"`)
	}
	if strings.TrimSpace(in.Category) == "" {
		return errors.New("category is required")
	}
	if _, err := time.Parse("2006-01-02", in.Date); err != nil {
		return errors.New("date must be in YYYY-MM-DD format")
	}
	return nil
}
