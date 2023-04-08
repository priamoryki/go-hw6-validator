package hw6validator

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

func NewValidationError(err error) ValidationError {
	return ValidationError{Err: err}
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	errs := make([]string, len(v))
	for i, err := range v {
		errs[i] = err.Err.Error()
	}
	return strings.Join(errs, "\n")
}

func Validate(v any) error {
	errs := ValidationErrors{}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		fieldInfo := val.Type().Field(i)
		tag := fieldInfo.Tag.Get("validate")

		// nested validation
		if fieldValue.Kind() == reflect.Struct {
			if err := Validate(fieldValue.Interface()); err != nil {
				errs = append(errs, NewValidationError(err))
			}
			continue
		}

		if !fieldInfo.IsExported() {
			if len(tag) != 0 {
				errs = append(errs, NewValidationError(ErrValidateForUnexportedFields))
			}
			continue
		}

		for _, tag := range strings.Split(tag, ";") {
			parsedTag := strings.Split(tag, ":")
			if len(parsedTag) != 2 || len(parsedTag[1]) == 0 {
				errs = append(errs, NewValidationError(ErrInvalidValidatorSyntax))
				continue
			}

			if err := fieldValidator(fieldValue, parsedTag[0], parsedTag[1]); err != nil {
				errs = append(errs, NewValidationError(err))
			}
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func fieldValidator(fieldValue reflect.Value, validator string, nonParsedArgs string) error {
	switch fieldValue.Kind() {
	case reflect.Int:
		value := int(fieldValue.Int())
		switch validator {
		case "in":
			return intInValidation(value, nonParsedArgs)
		case "min":
			return minValidation(value, nonParsedArgs)
		case "max":
			return maxValidation(value, nonParsedArgs)
		}
	case reflect.String:
		value := fieldValue.String()
		switch validator {
		case "len":
			return lenValidator(value, nonParsedArgs)
		case "in":
			return inValidation(value, strings.Split(nonParsedArgs, ","))
		case "min":
			return minValidation(len(value), nonParsedArgs)
		case "max":
			return maxValidation(len(value), nonParsedArgs)
		}
	}
	return nil
}

func stringsToInts(args []string) ([]int, error) {
	result := make([]int, len(args))
	for i := 0; i < len(result); i++ {
		a, err := strconv.Atoi(args[i])
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		result[i] = a
	}
	return result, nil
}

func lenValidator(value string, str string) error {
	length, err := strconv.Atoi(str)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	if len(value) != length {
		return errors.New("not valid value of struct with 'len' tag")
	}
	return nil
}

func inValidation[T comparable](value T, arr []T) error {
	for _, elem := range arr {
		if elem == value {
			return nil
		}
	}
	return errors.New("not valid value of struct with 'in' tag")
}

func intInValidation(value int, str string) error {
	args := strings.Split(str, ",")
	intArgs, err := stringsToInts(args)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	return inValidation(value, intArgs)
}

func minValidation(value int, str string) error {
	min, err := strconv.Atoi(str)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	if value < min {
		return errors.New("not valid value of struct with 'min' tag")
	}
	return nil
}

func maxValidation(value int, str string) error {
	max, err := strconv.Atoi(str)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}
	if value > max {
		return errors.New("not valid value of struct with 'max' tag")
	}
	return nil
}
