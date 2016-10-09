// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models


func countProfessors(e Engine) int64 {
	count, _ := e.Where("type=0").Count(new(User))
	return count
}

// CountProfessors returns number of users of type professor.
func CountProfessors() int64 {
	return countUsers(x)
}

// Users returns number of users in given page.
func Professors(page, pageSize int) ([]*User, error) {
	users := make([]*User, 0, pageSize)
	return users, x.Limit(pageSize, (page-1)*pageSize).Where("type=2").And("prohibit_login = ?", true).Asc("id").Find(&users)
}