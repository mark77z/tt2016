package auth

import (
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

type CreateNewCourseForm struct {
	Semester int64
	Group    int64
	Subject  int64
	Estatus  bool
}

func (f *CreateNewCourseForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}
