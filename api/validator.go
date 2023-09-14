package api

import (
	"github.com/go-playground/validator/v10"
)

const (
	USD = "USD"
	EUR = "EUR"
	BRL = "BRL"
)

var validCurrency validator.Func = func(fieldLvl validator.FieldLevel) bool {
	if currency, ok := fieldLvl.Field().Interface().(string); ok {
		return isValidCurrency(currency)
	}

	return false
}

func isValidCurrency(currency string) bool {
	switch currency {
	case USD, EUR, BRL:
		return true
	}

	return false
}
