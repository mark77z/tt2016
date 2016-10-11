// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models


func countApplications(e Engine) int64 {
	count, _ := e.Where("type=2").And("prohibit_login = ?", true).Count(new(User))
	return count
}

// CountApplications returns number of users of type professor.
func CountApplications() int64 {
	return countApplications(x)
}

// Users returns number of users in given page.
func Applications(page, pageSize int) ([]*User, error) {
	users := make([]*User, 0, pageSize)
	return users, x.Limit(pageSize, (page-1)*pageSize).Where("type=2").And("prohibit_login = ?", true).Asc("id").Find(&users)
}
