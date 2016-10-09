// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package subject

import (

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func Search(ctx *context.APIContext) {
	opts := &models.SearchSubjectOptions{
			Keyword:  ctx.Query("q"),
			OrderBy:  "Name DESC",
			PageSize: 10,
	}

	subjects, _, err := models.SearchSubjectByName(opts)
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	results := make([]*api.Subject, len(subjects))
	for i := range subjects {
		results[i] = &api.Subject{
			ID:        subjects[i].ID,
			Name:  	   subjects[i].Name,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}
