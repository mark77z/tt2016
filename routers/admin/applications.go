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
		Counter:  models.CountApplications,
		Ranger:   models.Applications,
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
	u.IsAdmin = false

	if err := models.UpdateUser(u); err != nil {
		ctx.Handle(500, "UpdateUser", err)
		return
	}
	log.Trace("Account profile updated by admin (%s): %s", ctx.User.Name, u.Name)

	models.SendApprovedAccountMail(ctx.Context, u)

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

	models.SendDeniedAccountMail(ctx.Context, u)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/applications",
	})
}
