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
type Tag struct {
	ID        	int64  `xorm:"pk autoincr"`
	Etiqueta    string `xorm:"VARCHAR(30) UNIQUE NOT NULL"`
}


// IsSubjectExist checks if given user name exist,
// the user name should be noncased unique.
// If uid is presented, then check will rule out that one,
// it is used when update a user name in settings page.
func IsTagExist(uid int64, name string) (bool, error) {
	if len(name) == 0 {
		return false, nil
	}
	return x.Where("id!=?", uid).Get(&Tag{Etiqueta: name})
}

var (
	reversedTagnames    = []string{"debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new", ".", ".."}
	reversedTagPatterns = []string{"*.keys"}
)

// isUsableName checks if name is reserved or pattern of name is not allowed
// based on given reversed names and patterns.
// Names are exact match, patterns can be prefix or suffix match with placeholder '*'.
func isUsableNameTag(names, patterns []string, name string) error {
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

func IsUsableTagName(name string) error {
	return isUsableNameTag(reversedTagnames, reversedTagPatterns, name)
}

// CreateTag creates record of a new tag.
func CreateTag(t *Tag) (err error) {
	if err = IsUsableTagName(t.Etiqueta); err != nil {
		return err
	}

	isExist, err := IsTagExist(0, t.Etiqueta)
	if err != nil {
		return err
	} else if isExist {
		return ErrTagAlreadyExist{t.Etiqueta}
		//return nil
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(t); err != nil {
		return err
	} 

	return sess.Commit()
}

func countTags(e Engine) int64 {
	count, _ := e.Where("id > 0").Count(new(Tag))
	return count
}

// CountGroups returns number of groups.
func CountTags() int64 {
	return countTags(x)
}

// Groups returns number of groups in given page.
func Tags(page, pageSize int) ([]*Tag, error) {
	tags := make([]*Tag, 0, pageSize)
	return tags, x.Limit(pageSize, (page-1)*pageSize).Where("id > 0").Asc("id").Find(&tags)
}


func getTags() ([]*Tag, error) {
	tags := make([]*Tag, 0, 5)
	return tags, x.Asc("etiqueta").Find(&tags)
}

func GetTags()([]*Tag, error){
	return getTags()
}


func updateTag(e Engine, t *Tag) error {
	// Organization does not need email
	if err := IsUsableTagName(t.Etiqueta); err != nil {
		return err
	}

	isExist, err := IsTagExist(0, t.Etiqueta)
	if err != nil {
		return err
	} else if isExist {
		return ErrTagAlreadyExist{t.Etiqueta}
		//return nil
	}

	_, err = e.Id(t.ID).AllCols().Update(t)
	return err
}

// UpdateSubject updates user's information.
func UpdateTag(t *Tag) error {
	return updateTag(x, t)
}


func deleteTag(e *xorm.Session, t *Tag) error {

	if _, err := e.Delete(&TagsRepo{TagID: t.ID}); err != nil {
		return err
	}

	if _, err := e.Id(t.ID).Delete(new(Tag)); err != nil {
		return fmt.Errorf("Delete: %v", err)
	}

	return nil
}

// DeleteGroup completely and permanently deletes everything of a user,
// but issues/comments/pulls will be kept and shown as someone has been deleted.
func DeleteTag(t *Tag) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteTag(sess, t); err != nil {
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}


func getTagByID(e Engine, id int64) (*Tag, error) {
	t := new(Tag)
	has, err := e.Id(id).Get(t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTagNotExist{id, ""}
	}
	return t, nil
}

// GetGroupByID returns the user object by given ID if exists.
func GetTagByID(id int64) (*Tag, error) {
	return getTagByID(x, id)
}

// GetSubjectByName returns user by given name.
func GetTagByName(name string) (*Tag, error) {
	if len(name) == 0 {
		return nil, ErrTagNotExist{0, name}
	}
	t := &Tag{Etiqueta: name}
	has, err := x.Get(t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTagNotExist{0, name}
	}
	return t, nil
}


// GetGroupIDsByNames returns a slice of ids corresponds to names.
func GetTagIDsByNames(names []string) []int64 {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		t, err := GetTagByName(name)
		if err != nil {
			continue
		}
		ids = append(ids, t.ID)
	}
	return ids
}

type SearchTagOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int // Can be smaller than or equal to setting.UI.ExplorePagingNum
}

// SearchSubjectByName takes keyword and part of user name to search,
// it returns results in given range and number of total results.
func SearchTagByName(opts *SearchTagOptions) (tags []*Tag, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return tags, 0, nil
	}
	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > setting.UI.ExplorePagingNum {
		opts.PageSize = setting.UI.ExplorePagingNum
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	tags = make([]*Tag, 0, opts.PageSize)
	// Append conditions
	sess := x.Where("etiqueta LIKE ?", searchQuery)

	var countSess xorm.Session
	countSess = *sess
	count, err := countSess.Count(new(Tag))
	if err != nil {
		return nil, 0, fmt.Errorf("Count: %v", err)
	}

	if len(opts.OrderBy) > 0 {
		sess.OrderBy(opts.OrderBy)
	}
	return tags, count, sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize).Find(&tags)
}
