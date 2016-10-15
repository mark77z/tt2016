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
	SUBJECTS  	 base.TplName = "admin/subject/list"
	SUBJECT_NEW  base.TplName = "admin/subject/new"
	SUBJECT_EDIT base.TplName = "admin/subject/edit"
)

func Subjects(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.subjects")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSubjects"] = true

	routers.RenderSubjectsSearch(ctx, &routers.SubjectsSearchOptions{
		Counter:  models.CountSubjects,
		Ranger:   models.Subjects,
		PageSize: setting.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  SUBJECTS,
	})
}

func NewSubject(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_subject")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSubjects"] = true
	ctx.HTML(200, SUBJECT_NEW)
}


func NewSubjectPost(ctx *context.Context, form auth.AdminCrateSubjectForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_subject")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSubjects"] = true

	if ctx.HasError() {
		ctx.HTML(200, SUBJECT_NEW)
		return
	}

	s := &models.Subject{
		Name:      form.Name,
	}

	if err := models.CreateSubject(s); err != nil {
		switch {
			case models.IsErrSubjectAlreadyExist(err):
				ctx.Data["Err_SubjectName"] = true
				ctx.RenderWithErr(ctx.Tr("form.subjectname_been_taken"), SUBJECT_NEW, &form)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_SubjectName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), SUBJECT_NEW, &form)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_SubjectName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SUBJECT_NEW, &form)
			default:
				ctx.Handle(500, "CreateSubject", err)
		}
		return
	}

	log.Trace("Subject created by admin (%s): %s", ctx.User.Name, s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.new_subject_success", s.Name))
	ctx.Redirect(setting.AppSubUrl + "/admin/subjects/new")
}

func prepareSubjectInfo(ctx *context.Context) *models.Subject {
	s, err := models.GetSubjectByID(ctx.ParamsInt64(":subjectid"))
	if err != nil {
		ctx.Handle(500, "GetSubjectByID", err)
		return nil
	}
	ctx.Data["Subject"] = s

	return s
}


func EditSubject(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_subject")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSubjects"] = true

	prepareSubjectInfo(ctx)
	if ctx.Written() {
		return
	}

	ctx.HTML(200, SUBJECT_EDIT)
}

func EditSubjectPost(ctx *context.Context, form auth.AdminEditSubjectForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_subject")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminSubjects"] = true

	s := prepareSubjectInfo(ctx)
	if ctx.Written() {
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, SUBJECT_EDIT)
		return
	}

	s.Name = form.Name

	if err := models.UpdateSubject(s); err != nil {
		if models.IsErrSubjectAlreadyExist(err) {
			ctx.Data["Err_SubjectName"] = true
			ctx.RenderWithErr(ctx.Tr("form.subjectname_been_taken"), SUBJECT_NEW, &form)
		} else {
			ctx.Handle(500, "UpdateSubject", err)
		}

		return
	}
	log.Trace("Subject updated by admin:%s", s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.update_subject_success"))
	ctx.Redirect(setting.AppSubUrl + "/admin/subjects/" + ctx.Params(":subjectid"))
}

func DeleteSubject(ctx *context.Context) {
	s, err := models.GetSubjectByID(ctx.ParamsInt64(":subjectid"))
	if err != nil {
		ctx.Handle(500, "GetSubjectByID", err)
		return
	}

	if err = models.DeleteSubject(s); err != nil {
		switch {
		default:
			ctx.Handle(500, "DeleteSubject", err)
		}
		return
	}
	log.Trace("Subject deleted by admin: %s", s.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_subject_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/subjects",
	})
}