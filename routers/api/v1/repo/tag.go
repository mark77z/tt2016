// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	api "github.com/gogits/go-gogs-client"

	"fmt"
	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func SearchTag(ctx *context.APIContext) {

	tags, err := models.GetTags()
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	results := make([]*api.Tag, len(tags))
	for i := range tags {
		results[i] = &api.Tag{
			ID:       tags[i].ID,
			Etiqueta: tags[i].Etiqueta,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}

func CreateTag(ctx *context.APIContext) {
	tag := ctx.Query("t")

	t := &models.Tag{
		Etiqueta: tag,
	}

	if err := models.CreateTag(t); err != nil {
		switch {
		case models.IsErrTagAlreadyExist(err):
			fmt.Errorf("TAG REPETIDA %v", err)
		default:
			fmt.Errorf("DEFAULT ERROR %v", err)
		}

		return
	}

	results := make([]*api.Tag, 1)
	results[0] = &api.Tag{
		ID:       t.ID,
		Etiqueta: tag,
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})

}
