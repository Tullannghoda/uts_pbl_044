package helper

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		switch e.Tag() {
		case "required":
			errors[field] = field + " wajib diisi"
		case "email":
			errors[field] = "Format email tidak valid"
		case "min":
			errors[field] = field + " minimal " + e.Param() + " karakter"
		case "max":
			errors[field] = field + " maksimal " + e.Param()
		case "len":
			errors[field] = field + " harus " + e.Param() + " karakter"
		case "numeric":
			errors[field] = field + " harus berupa angka"
		default:
			errors[field] = field + " tidak valid"
		}
	}
	return errors
}
