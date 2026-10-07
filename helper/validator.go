package helper

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

var tahunAkademikPattern = regexp.MustCompile(`^(\d{4})/(\d{4})-(Ganjil|Genap)$`)

func init() {
	validate = validator.New()

	// Pakai nama field sesuai tag json (bukan nama struct Go), supaya key di
	// response "errors" konsisten dengan nama field pada request body.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	// angkatan tidak boleh melebihi tahun berjalan (dihitung saat request masuk,
	// bukan dihardcode supaya tidak basi tahun depan).
	validate.RegisterValidation("max_current_year", func(fl validator.FieldLevel) bool {
		val := fl.Field().Int()
		return int(val) <= time.Now().Year()
	})

	// tahun_akademik wajib format "2026/2027-Ganjil" / "2026/2027-Genap",
	// dengan tahun kedua = tahun pertama + 1.
	validate.RegisterValidation("tahun_akademik", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		matches := tahunAkademikPattern.FindStringSubmatch(val)
		if matches == nil {
			return false
		}
		tahunAwal, err1 := strconv.Atoi(matches[1])
		tahunAkhir, err2 := strconv.Atoi(matches[2])
		if err1 != nil || err2 != nil {
			return false
		}
		return tahunAkhir == tahunAwal+1
	})
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
		case "max_current_year":
			errors[field] = field + " tidak boleh melebihi tahun berjalan (" + strconv.Itoa(time.Now().Year()) + ")"
		case "len":
			errors[field] = field + " harus " + e.Param() + " karakter"
		case "numeric":
			errors[field] = field + " harus berupa angka"
		case "tahun_akademik":
			errors[field] = field + " harus berformat seperti 2026/2027-Ganjil"
		default:
			errors[field] = field + " tidak valid"
		}
	}
	return errors
}