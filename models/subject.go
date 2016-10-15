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

	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/setting"
)

type Subject struct {
	ID   int64  `xorm:"pk autoincr"`
	Name string `xorm:"UNIQUE NOT NULL"`
}

func (s *Subject) ShortName(length int) string {
	return base.EllipsisString(s.Name, length)
}

// IsSubjectExist checks if given user name exist,
// the user name should be noncased unique.
// If uid is presented, then check will rule out that one,
// it is used when update a user name in settings page.
func IsSubjectExist(uid int64, name string) (bool, error) {
	if len(name) == 0 {
		return false, nil
	}
	return x.Where("id!=?", uid).Get(&Subject{Name: name})
}

var (
	reversedSubjectnames    = []string{"debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new", ".", ".."}
	reversedSubjectPatterns = []string{"*.keys"}
)

// isUsableName checks if name is reserved or pattern of name is not allowed
// based on given reversed names and patterns.
// Names are exact match, patterns can be prefix or suffix match with placeholder '*'.
func isUsableNameSubject(names, patterns []string, name string) error {
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

func IsUsableSubjectname(name string) error {
	return isUsableNameSubject(reversedSubjectnames, reversedSubjectPatterns, name)
}

// CreateSubject creates record of a new user.
func CreateSubject(s *Subject) (err error) {
	if err = IsUsableSubjectname(s.Name); err != nil {
		return err
	}

	isExist, err := IsSubjectExist(0, s.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrSubjectAlreadyExist{s.Name}
		//return nil
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(s); err != nil {
		return err
	}

	return sess.Commit()
}

func countSubjects(e Engine) int64 {
	count, _ := e.Where("id > 0").Count(new(Subject))
	return count
}

// CountSubjects returns number of subjects.
func CountSubjects() int64 {
	return countSubjects(x)
}

// Subjects returns number of subjects in given page.
func Subjects(page, pageSize int) ([]*Subject, error) {
	subjects := make([]*Subject, 0, pageSize)
	return subjects, x.Limit(pageSize, (page-1)*pageSize).Where("id > 0").Asc("id").Find(&subjects)
}

func getSubjects() ([]*Subject, error) {
	subjects := make([]*Subject, 0, 5)
	return subjects, x.Asc("name").Find(&subjects)
}

func GetSubjectsProfessor(ProfessorID int64) ([]*Subject, error) {
	subjects := make([]*Subject, 0, 5)
	return subjects, x.Cols("subject.id", "subject.name").Join("LEFT", "`course`", "`course`.subj_id=`subject`.id").Where("course.uid=?", ProfessorID).Asc("name").GroupBy("subject.id").Find(&subjects)
}

func GetSubjects() ([]*Subject, error) {
	return getSubjects()
}

func updateSubject(e Engine, s *Subject) error {
	if err := IsUsableSubjectname(s.Name); err != nil {
		return err
	}

	isExist, err := IsSubjectExist(0, s.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrSubjectAlreadyExist{s.Name}
		//return nil
	}

	_, err = e.Id(s.ID).AllCols().Update(s)
	return err
}

// UpdateSubject updates user's information.
func UpdateSubject(s *Subject) error {
	return updateSubject(x, s)
}

func deleteSubject(e *xorm.Session, s *Subject) error {

	if _, err := e.Id(s.ID).Delete(new(Subject)); err != nil {
		return fmt.Errorf("Delete: %v", err)
	}

	if err := deleteBeans(e,
		&Course{SubjID: s.ID},
	); err != nil {
		return fmt.Errorf("deleteBeans: %v", err)
	}

	return nil
}

// DeleteSubject completely and permanently deletes everything of a user,
// but issues/comments/pulls will be kept and shown as someone has been deleted.
func DeleteSubject(s *Subject) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteSubject(sess, s); err != nil {
		// Note: don't wrapper error here.
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}

func getSubjectByID(e Engine, id int64) (*Subject, error) {
	s := new(Subject)
	has, err := e.Id(id).Get(s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrSubjectNotExist{id, ""}
		//return nil,nil
	}
	return s, nil
}

// GetSubjectByID returns the subject object by given ID if exists.
func GetSubjectByID(id int64) (*Subject, error) {
	return getSubjectByID(x, id)
}

// GetSubjectByName returns subject by given name.
func GetSubjectByName(name string) (*Subject, error) {
	if len(name) == 0 {
		return nil, ErrSubjectNotExist{0, name}
		//return nil,nil
	}
	s := &Subject{Name: name}
	has, err := x.Get(s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrSubjectNotExist{0, name}
		//return nil,nil
	}
	return s, nil
}

// GetSubjectIDsByNames returns a slice of ids corresponds to names.
func GetSubjectIDsByNames(names []string) []int64 {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		s, err := GetSubjectByName(name)
		if err != nil {
			continue
		}
		ids = append(ids, s.ID)
	}
	return ids
}

type SearchSubjectOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int // Can be smaller than or equal to setting.UI.ExplorePagingNum
}

// SearchSubjectByName takes keyword and part of user name to search,
// it returns results in given range and number of total results.
func SearchSubjectByName(opts *SearchSubjectOptions) (subjects []*Subject, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return subjects, 0, nil
	}
	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > setting.UI.ExplorePagingNum {
		opts.PageSize = setting.UI.ExplorePagingNum
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	subjects = make([]*Subject, 0, opts.PageSize)
	// Append conditions
	sess := x.Where("name LIKE ?", searchQuery)

	var countSess xorm.Session
	countSess = *sess
	count, err := countSess.Count(new(Subject))
	if err != nil {
		return nil, 0, fmt.Errorf("Count: %v", err)
	}

	if len(opts.OrderBy) > 0 {
		sess.OrderBy(opts.OrderBy)
	}
	return subjects, count, sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize).Find(&subjects)
}
