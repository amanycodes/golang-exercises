package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type ValidationError struct {
	Field string
	Err   string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Field '%s': %s", e.Field, e.Err)
}

func ValidateStruct(s interface{}) []error {
	var errs []error
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("validate")

		if tag != "" {
			rules := parseTag(tag)

			if rules["required"] == "true" {
				if fieldVal.Kind() == reflect.String && fieldVal.String() == "" {
					errs = append(errs, ValidationError{Field: fieldType.Name, Err: "is required"})
				}
				if (fieldVal.Kind() == reflect.Int || fieldVal.Kind() == reflect.Int64 || fieldVal.Kind() == reflect.Int32) && fieldVal.Int() == 0 {
					errs = append(errs, ValidationError{Field: fieldType.Name, Err: "is required"})
				}
				if (fieldVal.Kind() == reflect.Float32 || fieldVal.Kind() == reflect.Float64) && fieldVal.Float() == 0 {
					errs = append(errs, ValidationError{Field: fieldType.Name, Err: "is required"})
				}
			}

			if pattern, ok := rules["regex"]; ok && fieldVal.Kind() == reflect.String {
				matched, err := regexp.MatchString(pattern, fieldVal.String())
				if err != nil || !matched {
					errs = append(errs, ValidationError{Field: fieldType.Name, Err: "does not match required pattern"})
				}
			}

			if minStr, ok := rules["min"]; ok {
				min, err := strconv.ParseFloat(minStr, 64)
				if err == nil {
					var value float64
					if fieldVal.Kind() == reflect.Int || fieldVal.Kind() == reflect.Int64 || fieldVal.Kind() == reflect.Int32 {
						value = float64(fieldVal.Int())
					} else if fieldVal.Kind() == reflect.Float32 || fieldVal.Kind() == reflect.Float64 {
						value = fieldVal.Float()
					}
					if value < min {
						errs = append(errs, ValidationError{Field: fieldType.Name, Err: fmt.Sprintf("should be at least %v", min)})
					}
				}
			}

			if maxStr, ok := rules["max"]; ok {
				max, err := strconv.ParseFloat(maxStr, 64)
				if err == nil {
					var value float64
					if fieldVal.Kind() == reflect.Int || fieldVal.Kind() == reflect.Int64 || fieldVal.Kind() == reflect.Int32 {
						value = float64(fieldVal.Int())
					} else if fieldVal.Kind() == reflect.Float32 || fieldVal.Kind() == reflect.Float64 {
						value = fieldVal.Float()
					}
					if value > max {
						errs = append(errs, ValidationError{Field: fieldType.Name, Err: fmt.Sprintf("should be at most %v", max)})
					}
				}
			}
		}

		if fieldVal.Kind() == reflect.Struct {
			nestedErrs := ValidateStruct(fieldVal.Interface())
			for _, err := range nestedErrs {
				if vErr, ok := err.(ValidationError); ok {
					errs = append(errs, ValidationError{Field: fieldType.Name + "." + vErr.Field, Err: vErr.Err})
				} else {
					errs = append(errs, err)
				}
			}
		}
	}
	return errs
}

func parseTag(tag string) map[string]string {
	rules := make(map[string]string)
	parts := splitComma(tag)
	for _, part := range parts {
		if part == "required" {
			rules["required"] = "true"
		} else if kv := splitEqual(part); len(kv) == 2 {
			rules[kv[0]] = kv[1]
		}
	}
	return rules
}

func splitComma(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitEqual(s string) []string {
	var parts []string
	current := ""
	for i, c := range s {
		if c == '=' {
			parts = append(parts, current)
			parts = append(parts, s[i+1:])
			return parts
		}
		current += string(c)
	}
	return []string{s}
}

type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
}

type Person struct {
	Name    string  `validate:"required,regex=^[A-Za-z]+$"`
	Age     int     `validate:"min=18,max=60"`
	Address Address `validate:"required"`
}

func main() {
	p := Person{
		Name: "",
		Age:  16,
		Address: Address{
			Street: "",
			City:   "New York",
		},
	}
	errors := ValidateStruct(p)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("Validation passed!")
	}
}
