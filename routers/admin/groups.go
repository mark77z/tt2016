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
	GROUPS  	  base.TplName  = "admin/group/list"
	GROUP_NEW  	  base.TplName = "admin/group/new"
	GROUP_EDIT    base.TplName = "admin/group/edit"
)

func Groups(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.groups")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminGroups"] = true

	routers.RenderGroupsSearch(ctx, &routers.GrupsSearchOptions{
		Counter:  models.CountGroups,
		Ranger:   models.Groups,
		PageSize: setting.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  GROUPS,
	})
}

func NewGroup(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_group")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminGroups"] = true
	ctx.HTML(200, GROUP_NEW)
}


func NewGroupPost(ctx *context.Context, form auth.AdminCreateGroupForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.new_group")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminGroups"] = true

	if ctx.HasError() {
		ctx.HTML(200, GROUP_NEW)
		return
	}

	g := &models.Group{
		Name: form.Name,
	}

	if err := models.CreateGroup(g); err != nil {
		switch {
			case models.IsErrGroupAlreadyExist(err):
				ctx.Data["Err_GroupName"] = true
				ctx.RenderWithErr(ctx.Tr("form.groupname_been_taken"), GROUP_NEW, &form)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_SemesterName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), GROUP_NEW, &form)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_SemesterName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), GROUP_NEW, &form)
			default:
				ctx.Handle(500, "CreateGroup", err)
		}
		return
	}

	log.Trace("Group created by admin: %s", g.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.new_group_success", g.Name))
	ctx.Redirect(setting.AppSubUrl + "/admin/groups/new")
}

func prepareGroupInfo(ctx *context.Context) *models.Group {
	g, err := models.GetGroupByID(ctx.ParamsInt64(":groupid"))
	if err != nil {
		ctx.Handle(500, "GetGroupByID", err)
		return nil
	}
	ctx.Data["Group"] = g

	return g
}


func EditGroup(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_group")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminGroups"] = true

	prepareGroupInfo(ctx)
	if ctx.Written() {
		return
	}

	ctx.HTML(200, GROUP_EDIT)
}

func EditGroupPost(ctx *context.Context, form auth.AdminEditGroupForm) {
	ctx.Data["Title"] = ctx.Tr("admin.users.edit_group")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminGroups"] = true

	g := prepareGroupInfo(ctx)
	if ctx.Written() {
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, GROUP_EDIT)
		return
	}

	g.Name = form.Name

	if err := models.UpdateGroup(g); err != nil {
		if models.IsErrGroupAlreadyExist(err) {
			ctx.Data["Err_GroupName"] = true
			ctx.RenderWithErr(ctx.Tr("form.groupname_been_taken"), GROUP_NEW, &form)
		} else {
			ctx.Handle(500, "UpdateGroup", err)
		}

		return
	}
	log.Trace("Group updated by admin:%s", g.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.update_group_success"))
	ctx.Redirect(setting.AppSubUrl + "/admin/groups/" + ctx.Params(":groupid"))
}

func DeleteGroup(ctx *context.Context) {
	g, err := models.GetGroupByID(ctx.ParamsInt64(":groupid"))
	if err != nil {
		ctx.Handle(500, "GetGroupByID", err)
		return
	}

	if err = models.DeleteGroup(g); err != nil {
		switch {
			default:
				ctx.Handle(500, "DeleteGroup", err)
			}
		return
	}
	log.Trace("Group deleted by admin: %s", g.Name)

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_group_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/groups",
	})
}