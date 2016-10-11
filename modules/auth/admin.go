// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"gopkg.in/macaron.v1"

	"github.com/go-macaron/binding"
)

type AdminCrateUserForm struct {
	LoginType  string `binding:"Required"`
	LoginName  string
	UserName   string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	FullName   string `binding:"Required;MaxSize(60)"`
	Email      string `binding:"Required;Email;MaxSize(254)"`
	Password   string `binding:"MaxSize(255)"`
	SendNotify bool
}

func (f *AdminCrateUserForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type AdminEditUserForm struct {
	LoginType        string `binding:"Required"`
	LoginName        string
	FullName         string `binding:"MaxSize(100)"`
	Email            string `binding:"Required;Email;MaxSize(254)"`
	Password         string `binding:"MaxSize(255)"`
	Website          string `binding:"MaxSize(50)"`
	Location         string `binding:"MaxSize(50)"`
	MaxRepoCreation  int
	Active           bool
	Admin            bool
	AllowGitHook     bool
	AllowImportLocal bool
	ProhibitLogin    bool
}

func (f *AdminEditUserForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type AdminCrateSubjectForm struct {
	Name 	string `binding:"Required;MaxSize(90)"`
}

type AdminEditSubjectForm struct {
	Name    string `binding:"Required;MaxSize(90)"`
}

func (f *AdminCrateSubjectForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type AdminCrateSemesterForm struct {
	Name 	string `binding:"Required;MaxSize(90)"`
}

type AdminEditSemesterForm struct {
	Name    string `binding:"Required;MaxSize(90)"`
}

func (f *AdminCrateSemesterForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type AdminCreateGroupForm struct {
	Name 	string `binding:"Required;MaxSize(6)"`
}

type AdminEditGroupForm struct {
	Name    string `binding:"Required;MaxSize(6)"`
}

func (f *AdminCreateGroupForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

//*************TAG FORMS***************
type AdminCreateTagForm struct {
	Etiqueta	string `binding:"Required;MaxSize(50)"`
}

type AdminEditTagForm struct {
	Etiqueta    string `binding:"Required;MaxSize(50)"`
}

func (f *AdminCreateTagForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}