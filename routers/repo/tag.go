// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
)

const (
	CREATE_TAG  base.TplName = "tag/create"
)

func CreateTag(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("new_tag")
	ctx.HTML(200, CREATE_TAG)
}

func CreateTagPost(ctx *context.Context, form auth.CreateTagForm) {
	ctx.Data["Title"] = ctx.Tr("new_tag")

	t := &models.Tag{
		Etiqueta: form.Etiqueta,
	}

	if err := models.CreateTag(t); err != nil {
		switch {
			case models.IsErrTagAlreadyExist(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("form.tagname_been_taken"), CREATE_TAG, &form)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), CREATE_TAG, &form)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_TagName"] = true
				ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), CREATE_TAG, &form)
			default:
				ctx.Handle(500, "CreateGroup", err)
		}
		return
	}

	log.Trace("Tag created: %s", t.Etiqueta)

	ctx.Flash.Success(ctx.Tr("admin.new_tag_success", t.Etiqueta))
	ctx.Redirect(setting.AppSubUrl + "/tag/create")
}
