// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	//"container/list"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/go-xorm/xorm"

	//"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/setting"
)

// Subject represents the object of individual and member of organization.
type Group struct {
	ID   int64  `xorm:"pk autoincr"`
	Name string `xorm:"VARCHAR(6) UNIQUE NOT NULL"`
}

// IsSubjectExist checks if given user name exist,
// the user name should be noncased unique.
// If uid is presented, then check will rule out that one,
// it is used when update a user name in settings page.
func IsGroupExist(uid int64, name string) (bool, error) {
	if len(name) == 0 {
		return false, nil
	}
	return x.Where("id!=?", uid).Get(&Group{Name: name})
}

var (
	reversedGroupnames    = []string{"debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new", ".", ".."}
	reversedGroupPatterns = []string{"*.keys"}
)

// isUsableName checks if name is reserved or pattern of name is not allowed
// based on given reversed names and patterns.
// Names are exact match, patterns can be prefix or suffix match with placeholder '*'.
func isUsableNameGroup(names, patterns []string, name string) error {
	name = strings.TrimSpace(strings.ToLower(name))
	if utf8.RuneCountInString(name) == 0 {
		return ErrNameEmpty
	}

	for i := range names {
		if name == names[i] {
			return ErrNameReserved{name}
		}
	}

	for _, pat := range patterns {
		if pat[0] == '*' && strings.HasSuffix(name, pat[1:]) ||
			(pat[len(pat)-1] == '*' && strings.HasPrefix(name, pat[:len(pat)-1])) {
			return ErrNamePatternNotAllowed{pat}
		}
	}

	return nil
}

func IsUsableGroupname(name string) error {
	return isUsableNameGroup(reversedGroupnames, reversedGroupPatterns, name)
}

// CreateSubject creates record of a new user.
func CreateGroup(g *Group) (err error) {
	if err = IsUsableGroupname(g.Name); err != nil {
		return err
	}

	isExist, err := IsGroupExist(0, g.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrGroupAlreadyExist{g.Name}
		//return nil
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(g); err != nil {
		return err
	}

	return sess.Commit()
}

func countGroups(e Engine) int64 {
	count, _ := e.Where("id > 0").Count(new(Group))
	return count
}

// CountGroups returns number of groups.
func CountGroups() int64 {
	return countGroups(x)
}

// Groups returns number of groups in given page.
func Groups(page, pageSize int) ([]*Group, error) {
	groups := make([]*Group, 0, pageSize)
	return groups, x.Limit(pageSize, (page-1)*pageSize).Where("id > 0").Asc("id").Find(&groups)
}

func getGroups() ([]*Group, error) {
	groups := make([]*Group, 0, 5)
	return groups, x.Asc("name").Find(&groups)
}

func GetGroups() ([]*Group, error) {
	return getGroups()
}

func updateGroup(e Engine, g *Group) error {
	// Organization does not need email
	if err := IsUsableGroupname(g.Name); err != nil {
		return err
	}

	isExist, err := IsGroupExist(0, g.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrGroupAlreadyExist{g.Name}
		//return nil
	}

	_, err = e.Id(g.ID).AllCols().Update(g)
	return err
}

// UpdateSubject updates user's information.
func UpdateGroup(g *Group) error {
	return updateGroup(x, g)
}

func deleteGroup(e *xorm.Session, g *Group) error {

	if _, err := e.Id(g.ID).Delete(new(Group)); err != nil {
		return fmt.Errorf("Delete: %v", err)
	}

	return nil
}

// DeleteGroup completely and permanently deletes everything of a user,
// but issues/comments/pulls will be kept and shown as someone has been deleted.
func DeleteGroup(g *Group) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteGroup(sess, g); err != nil {
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}

func getGroupByID(e Engine, id int64) (*Group, error) {
	g := new(Group)
	has, err := e.Id(id).Get(g)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrGroupNotExist{id, ""}
	}
	return g, nil
}

// GetGroupByID returns the user object by given ID if exists.
func GetGroupByID(id int64) (*Group, error) {
	return getGroupByID(x, id)
}

// GetSubjectByName returns user by given name.
func GetGroupByName(name string) (*Group, error) {
	if len(name) == 0 {
		return nil, ErrGroupNotExist{0, name}
	}
	g := &Group{Name: name}
	has, err := x.Get(g)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrGroupNotExist{0, name}
	}
	return g, nil
}

// GetGroupIDsByNames returns a slice of ids corresponds to names.
func GetGroupIDsByNames(names []string) []int64 {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		g, err := GetGroupByName(name)
		if err != nil {
			continue
		}
		ids = append(ids, g.ID)
	}
	return ids
}

type SearchGroupOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int // Can be smaller than or equal to setting.UI.ExplorePagingNum
}

// SearchSubjectByName takes keyword and part of user name to search,
// it returns results in given range and number of total results.
func SearchGroupByName(opts *SearchGroupOptions) (groups []*Group, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return groups, 0, nil
	}
	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > setting.UI.ExplorePagingNum {
		opts.PageSize = setting.UI.ExplorePagingNum
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	groups = make([]*Group, 0, opts.PageSize)
	// Append conditions
	sess := x.Where("name LIKE ?", searchQuery)

	var countSess xorm.Session
	countSess = *sess
	count, err := countSess.Count(new(Group))
	if err != nil {
		return nil, 0, fmt.Errorf("Count: %v", err)
	}

	if len(opts.OrderBy) > 0 {
		sess.OrderBy(opts.OrderBy)
	}
	return groups, count, sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize).Find(&groups)
}

func GetGroupsProfessor(ProfessorID int64) ([]*Group, error) {
	groups := make([]*Group, 0, 5)
	return groups, x.Cols("group.id", "group.name").Join("LEFT", "`course`", "`course`.group_id=`group`.id").Where("course.uid=?", ProfessorID).Asc("name").GroupBy("group.id").Find(&groups)
}
