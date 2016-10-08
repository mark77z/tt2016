// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	//"strings"

	//"github.com/Unknwon/com"

	"github.com/gogits/gogs/models"
	//"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
	"github.com/gogits/gogs/routers"
)

const (
	APPLICATIONS     base.TplName = "admin/application/list"
)

func Applications(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.applications")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminApplications"] = true

	routers.RenderUserSearch(ctx, &routers.UserSearchOptions{
		Type:     models.USER_TYPE_PROFESSOR,
		Counter:  models.CountUsers,
		Ranger:   models.Professors,
		PageSize: setting.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  APPLICATIONS,
	})
}


func ActivateUser(ctx *context.Context) {
	u, err := models.GetUserByID(ctx.QueryInt64("id"))
	if err != nil {
		ctx.Handle(500, "GetUserByID", err)
		return
	}

	u.IsActive = true
	u.ProhibitLogin = false

	if err := models.UpdateUser(u); err != nil {
		ctx.Handle(500, "UpdateUser", err)
		return
	}
	log.Trace("Account profile updated by admin (%s): %s", ctx.User.Name, u.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.update_profile_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/applications",
	})
}

func DeleteApplicationUser(ctx *context.Context) {
	u, err := models.GetUserByID(ctx.QueryInt64("id"))
	if err != nil {
		ctx.Handle(500, "GetUserByID", err)
		return
	}

	if err = models.DeleteUser(u); err != nil {
		switch {
		case models.IsErrUserOwnRepos(err):
			ctx.Flash.Error(ctx.Tr("admin.users.still_own_repo"))
			ctx.JSON(200, map[string]interface{}{
				"redirect": setting.AppSubUrl + "/admin/users/" + ctx.Params(":userid"),
			})
		case models.IsErrUserHasOrgs(err):
			ctx.Flash.Error(ctx.Tr("admin.users.still_has_org"))
			ctx.JSON(200, map[string]interface{}{
				"redirect": setting.AppSubUrl + "/admin/users/" + ctx.Params(":userid"),
			})
		default:
			ctx.Handle(500, "DeleteUser", err)
		}
		return
	}
	log.Trace("Account deleted by admin (%s): %s", ctx.User.Name, u.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/applications",
	})
}

/* 
func EditUserPost(ctx *context.Context, form auth.AdminEditUserForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_account")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminUsers"] = true

	u := prepareUserInfo(ctx)
	if ctx.Written() {
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, USER_EDIT)
		return
	}

	fields := strings.Split(form.LoginType, "-")
	if len(fields) == 2 {
		loginType := models.LoginType(com.StrTo(fields[0]).MustInt())
		loginSource := com.StrTo(fields[1]).MustInt64()

		if u.LoginSource != loginSource {
			u.LoginSource = loginSource
			u.LoginType = loginType
		}
	}

	if len(form.Password) > 0 {
		u.Passwd = form.Password
		u.Salt = models.GetUserSalt()
		u.EncodePasswd()
	}

	u.LoginName = form.LoginName
	u.FullName = form.FullName
	u.Email = form.Email
	u.Website = form.Website
	u.Location = form.Location
	u.MaxRepoCreation = form.MaxRepoCreation
	u.IsActive = form.Active
	u.IsAdmin = form.Admin
	u.AllowGitHook = form.AllowGitHook
	u.AllowImportLocal = form.AllowImportLocal
	u.ProhibitLogin = form.ProhibitLogin

	if err := models.UpdateUser(u); err != nil {
		if models.IsErrEmailAlreadyUsed(err) {
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(ctx.Tr("form.email_been_used"), USER_EDIT, &form)
		} else {
			ctx.Handle(500, "UpdateUser", err)
		}
		return
	}
	log.Trace("Account profile updated by admin (%s): %s", ctx.User.Name, u.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.update_profile_success"))
	ctx.Redirect(setting.AppSubUrl + "/admin/users/" + ctx.Params(":userid"))
}

func DeleteUser(ctx *context.Context) {
	u, err := models.GetUserByID(ctx.ParamsInt64(":userid"))
	if err != nil {
		ctx.Handle(500, "GetUserByID", err)
		return
	}

	if err = models.DeleteUser(u); err != nil {
		switch {
		case models.IsErrUserOwnRepos(err):
			ctx.Flash.Error(ctx.Tr("admin.users.still_own_repo"))
			ctx.JSON(200, map[string]interface{}{
				"redirect": setting.AppSubUrl + "/admin/users/" + ctx.Params(":userid"),
			})
		case models.IsErrUserHasOrgs(err):
			ctx.Flash.Error(ctx.Tr("admin.users.still_has_org"))
			ctx.JSON(200, map[string]interface{}{
				"redirect": setting.AppSubUrl + "/admin/users/" + ctx.Params(":userid"),
			})
		default:
			ctx.Handle(500, "DeleteUser", err)
		}
		return
	}
	log.Trace("Account deleted by admin (%s): %s", ctx.User.Name, u.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/users",
	})
}
*/