package auth

import (
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

type CreateTagForm struct {
	Etiqueta		string
}

func (f *CreateTagForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}