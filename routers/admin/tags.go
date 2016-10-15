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
	TAG  	  	base.TplName  = "admin/tag/list"
	TAG_NEW  	base.TplName = "admin/tag/new"
	TAG_EDIT   base.TplName = "admin/tag/edit"
)

func Tags(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.tags")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminTags"] = true

	routers.RenderTagsSearch(ctx, &routers.TagsSearchOptions{
		Counter:  models.CountTags,
		Ranger:   models.Tags,
		PageSize: setting.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  TAG,
	})
}

func NewTag(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.new_tag")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminTags"] = true
	ctx.HTML(200, TAG_NEW)
}


func NewTagPost(ctx *context.Context, form auth.AdminCreateTagForm) {
	ctx.Data["Title"] = ctx.Tr("admin.new_tag")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminTags"] = true

	if ctx.HasError() {
		ctx.HTML(200, TAG_NEW)
		return
	}

	t := &models.Tag{
		Etiqueta: form.Etiqueta,
	}

	if err := models.CreateTag(t); err != nil {
		switch {
			case models.IsErrTagAlreadyExist(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("form.tagname_been_taken"), TAG_NEW, &form)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), TAG_NEW, &form)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), TAG_NEW, &form)
			default:
				ctx.Handle(500, "CreateGroup", err)
		}
		return
	}

	log.Trace("Tag created by admin: %s", t.Etiqueta)

	ctx.Flash.Success(ctx.Tr("admin.new_tag_success", t.Etiqueta))
	ctx.Redirect(setting.AppSubUrl + "/admin/tags/new")
}

func prepareTagInfo(ctx *context.Context) *models.Tag {
	t, err := models.GetTagByID(ctx.ParamsInt64(":tagid"))
	if err != nil {
		ctx.Handle(500, "GetTagByID", err)
		return nil
	}
	ctx.Data["Tag"] = t

	return t
}


func EditTag(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.edit_tag")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminTags"] = true

	prepareTagInfo(ctx)
	if ctx.Written() {
		return
	}

	ctx.HTML(200, TAG_EDIT)
}

func EditTagPost(ctx *context.Context, form auth.AdminEditTagForm) {
	ctx.Data["Title"] = ctx.Tr("admin.edit_tag")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminTags"] = true

	t := prepareTagInfo(ctx)
	if ctx.Written() {
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, TAG_EDIT)
		return
	}

	t.Etiqueta = form.Etiqueta

	if err := models.UpdateTag(t); err != nil {
		if models.IsErrTagAlreadyExist(err) {
			ctx.Data["Err_TagName"] = true
			ctx.RenderWithErr(ctx.Tr("form.tagname_been_taken"), TAG_NEW, &form)
		} else {
			ctx.Handle(500, "UpdateTag", err)
		}

		return
	}
	log.Trace("Tag updated by admin:%s", t.Etiqueta)

	ctx.Flash.Success(ctx.Tr("admin.update_tag_success"))
	ctx.Redirect(setting.AppSubUrl + "/admin/tags/" + ctx.Params(":tagid"))
}

func DeleteTag(ctx *context.Context) {
	t, err := models.GetTagByID(ctx.ParamsInt64(":tagid"))
	if err != nil {
		ctx.Handle(500, "GetTagByID", err)
		return
	}

	if err = models.DeleteTag(t); err != nil {
		switch {
			default:
				ctx.Handle(500, "DeleteTag", err)
			}
		return
	}
	log.Trace("Tag deleted by admin: %s", t.Etiqueta)

	ctx.Flash.Success(ctx.Tr("admin.deletion_tag_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/admin/tags",
	})
}