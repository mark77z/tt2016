// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	//"strings"

	//"github.com/Unknwon/com"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
	"github.com/gogits/gogs/routers"
)

const (
	SEMESTERS  	 base.TplName  = "admin/semester/list"
	SEMESTER_NEW  base.TplName = "admin/semester/new"
	SEMESTER_EDIT base.TplName = "admin/semester/edit"
)

func Semesters(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.semesters")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSemesters"] = true

	routers.RenderSemestersSearch(ctx, &routers.SemestersSearchOptions{
		Counter:  models.CountSemesters,
		Ranger:   models.Semesters,
		PageSize: setting.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  SEMESTERS,
	})
}

func NewSemester(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_semester")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSemesters"] = true
	ctx.HTML(200, SEMESTER_NEW)
}


func NewSemesterPost(ctx *context.Context, form auth.AdminCrateSemesterForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_semester")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSemesters"] = true

	if ctx.HasError() {
		ctx.HTML(200, SEMESTER_NEW)
		return
	}

	s := &models.Semester{
		Name:      form.Name,
	}

	if err := models.CreateSemester(s); err != nil {
		switch {
			case models.IsErrSemesterAlreadyExist(err):
				ctx.Data["Err_SemesterName"] = true
				ctx.RenderWithErr(ctx.Tr("form.semestername_been_taken"), SEMESTER_NEW, &form)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_SemesterName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), SEMESTER_NEW, &form)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_SemesterName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SEMESTER_NEW, &form)
			default:
				ctx.Handle(500, "CreateSemester", err)
		}
		return
	}

	log.Trace("Semester created by admin: %s", s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.new_semester_success", s.Name))
	ctx.Redirect(setting.AppSubUrl + "/admin/semesters/new")
}

func prepareSemesterInfo(ctx *context.Context) *models.Semester {
	s, err := models.GetSemesterByID(ctx.ParamsInt64(":semesterid"))
	if err != nil {
		ctx.Handle(500, "GetSemesterByID", err)
		return nil
	}
	ctx.Data["Semester"] = s

	return s
}


func EditSemester(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_semester")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSemesters"] = true

	prepareSemesterInfo(ctx)
	if ctx.Written() {
		return
	}

	ctx.HTML(200, SEMESTER_EDIT)
}

func EditSemesterPost(ctx *context.Context, form auth.AdminEditSemesterForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_semester")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSemesters"] = true

	s := prepareSemesterInfo(ctx)
	if ctx.Written() {
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, SEMESTER_EDIT)
		return
	}

	s.Name = form.Name

	if err := models.UpdateSemester(s); err != nil {
		if models.IsErrSemesterAlreadyExist(err) {
			ctx.Data["Err_SemesterName"] = true
			ctx.RenderWithErr(ctx.Tr("form.semestername_been_taken"), SEMESTER_NEW, &form)
		} else {
			ctx.Handle(500, "UpdateSemester", err)
		}

		return
	}
	log.Trace("Semester updated by admin:%s", s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.update_semester_success"))
	ctx.Redirect(setting.AppSubUrl + "/admin/semesters/" + ctx.Params(":semesterid"))
}

func DeleteSemester(ctx *context.Context) {
	s, err := models.GetSemesterByID(ctx.ParamsInt64(":semesterid"))
	if err != nil {
		ctx.Handle(500, "GetSemesterByID", err)
		return
	}

	if err = models.DeleteSemester(s); err != nil {
		switch {
		default:
			ctx.Handle(500, "DeleteSemester", err)
		}
		return
	}
	log.Trace("Semester deleted by admin: %s", s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_semester_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/semesters",
	})
}