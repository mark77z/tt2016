// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package semester

import (
	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
	"strconv"
)

func SearchByProfessor(ctx *context.APIContext) {
	ProfessorID, _ := strconv.ParseInt(ctx.Query("q"), 10, 64)
	semesters, err := models.GetSemestersProfessor(ProfessorID)
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	results := make([]*api.Semester, len(semesters))
	for i := range semesters {
		results[i] = &api.Semester{
			ID:   semesters[i].ID,
			Name: semesters[i].Name,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}
