package lib

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type AppError struct {
	Message string `json:"message" yaml:"message" xml:"message"`
	Code    int    `json:"code" yaml:"code" xml:"code"`
}

func (ae AppError) Error() string {
	return fmt.Sprintf("Error(%d): %s", ae.Code, ae.Message)
}

func InvalidArgError(name string, value string, options []string, code int) AppError {
	return AppError{
		Message: fmt.Sprintf("Invalid argument %s: %s. Valid options are: %s", name, value, strings.Join(options, ", ")),
		Code:    code,
	}
}

func UnresolvablePathError(fsPath string) AppError {
	return AppError{
		Message: fmt.Sprintf("Unable to resolve file system path %s", fsPath),
		Code:    UnresolvableFsPath,
	}
}

const (
	InvalidFormatCode = iota + 400
	InvalidLevelCode
	InvalidColorCode
	UnresolvableFsPath
)

func handleStopCode(err error) {
	if err != nil {
		appErr := new(AppError)
		if ok := errors.As(err, &appErr); ok {
			os.Exit(appErr.Code)
		} else {
			os.Exit(1)
		}
	}
}

func HandleStopError(err error) {
	if err != nil {
		fmt.Println(err)
		handleStopCode(err)
	}
}
