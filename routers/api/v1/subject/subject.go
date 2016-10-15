// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package subject

import (
	api "github.com/gogits/go-gogs-client"

	"fmt"
	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
	"strconv"
)

func Search(ctx *context.APIContext) {
	opts := &models.SearchSubjectOptions{
		Keyword:  ctx.Query("q"),
		OrderBy:  "Name DESC",
		PageSize: 10,
	}

	longitud := len(ctx.Query("q"))
	subjects := make([]*models.Subject, 0, 10)
	var err error

	if longitud < 2 {
		subjects, err = models.GetSubjects()
		if err != nil {
			ctx.JSON(500, map[string]interface{}{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}
	} else {
		subjects, _, err = models.SearchSubjectByName(opts)
		if err != nil {
			ctx.JSON(500, map[string]interface{}{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}
	}

	results := make([]*api.Subject, len(subjects))
	for i := range subjects {
		results[i] = &api.Subject{
			ID:   subjects[i].ID,
			Name: subjects[i].Name,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}

func List(ctx *context.APIContext) {
	subjects, err := models.GetSubjects()
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
			ID:   subjects[i].ID,
			Name: subjects[i].Name,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}

func SearchByProfessor(ctx *context.APIContext) {
	ProfessorID, _ := strconv.ParseInt(ctx.Query("q"), 10, 64)
	subjects, err := models.GetSubjectsProfessor(ProfessorID)
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
			ID:   subjects[i].ID,
			Name: subjects[i].Name,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}

func CreateSubject(ctx *context.APIContext) {
	subject := ctx.Query("s")

	s := &models.Subject{
		Name: subject,
	}

	if err := models.CreateSubject(s); err != nil {
		switch {
		case models.IsErrSubjectAlreadyExist(err):
			fmt.Errorf("MATERIA REPETIDA %v", err)
		default:
			fmt.Errorf("DEFAULT ERROR %v", err)
		}

		return
	}

	results := make([]*api.Subject, 1)
	results[0] = &api.Subject{
		ID:   s.ID,
		Name: subject,
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})

}
