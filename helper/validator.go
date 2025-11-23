package helper

import "go-fiber/app/model"

func ValidateLoginRequest(req *model.LoginRequest) []string {
    var errs []string

    if req.Username == "" {
        errs = append(errs, "username required")
    }

    if req.Password == "" {
        errs = append(errs, "password required")
    }

    return errs
}
