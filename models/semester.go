// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/go-xorm/xorm"

	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/setting"
)

// Semester represents the object of individual and member of organization.
type Semester struct {
	ID   int64  `xorm:"pk autoincr"`
	Name string `xorm:"VARCHAR(90) UNIQUE NOT NULL"`
}

func (s *Semester) ShortName(length int) string {
	return base.EllipsisString(s.Name, length)
}

// IsSemesterExist checks if given user name exist,
// the user name should be noncased unique.
// If uid is presented, then check will rule out that one,
// it is used when update a user name in settings page.
func IsSemesterExist(uid int64, name string) (bool, error) {
	if len(name) == 0 {
		return false, nil
	}
	return x.Where("id!=?", uid).Get(&Semester{Name: name})
}

var (
	reversedSemesternames    = []string{"debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new", ".", ".."}
	reversedSemesterPatterns = []string{"*.keys"}
)

// isUsableName checks if name is reserved or pattern of name is not allowed
// based on given reversed names and patterns.
// Names are exact match, patterns can be prefix or suffix match with placeholder '*'.
func isUsableNameSemester(names, patterns []string, name string) error {
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

func IsUsableSemestername(name string) error {
	return isUsableNameSemester(reversedSemesternames, reversedSemesterPatterns, name)
}

// CreateSemester creates record of a new user.
func CreateSemester(s *Semester) (err error) {
	if err = IsUsableSemestername(s.Name); err != nil {
		return err
	}

	isExist, err := IsSemesterExist(0, s.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrSemesterAlreadyExist{s.Name}
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

func countSemesters(e Engine) int64 {
	count, _ := e.Where("1").Count(new(Semester))
	return count
}

// CountSemesters returns number of semesters.
func CountSemesters() int64 {
	return countSemesters(x)
}

// Semesters returns number of semesters in given page.
func Semesters(page, pageSize int) ([]*Semester, error) {
	semesters := make([]*Semester, 0, pageSize)
	return semesters, x.Limit(pageSize, (page-1)*pageSize).Asc("id").Find(&semesters)
}

func updateSemester(e Engine, s *Semester) error {
	// Organization does not need email
	if err := IsUsableSemestername(s.Name); err != nil {
		return err
	}

	isExist, err := IsSemesterExist(0, s.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrSemesterAlreadyExist{s.Name}
		//return nil
	}

	_, err = e.Id(s.ID).AllCols().Update(s)
	return err
}

// UpdateSemester updates user's information.
func UpdateSemester(s *Semester) error {
	return updateSemester(x, s)
}

func deleteSemester(e *xorm.Session, s *Semester) error {

	if _, err := e.Id(s.ID).Delete(new(Semester)); err != nil {
		return fmt.Errorf("Delete: %v", err)
	}

	return nil
}

// DeleteSemester completely and permanently deletes everything of a user,
// but issues/comments/pulls will be kept and shown as someone has been deleted.
func DeleteSemester(s *Semester) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteSemester(sess, s); err != nil {
		// Note: don't wrapper error here.
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}

func getSemesterByID(e Engine, id int64) (*Semester, error) {
	s := new(Semester)
	has, err := e.Id(id).Get(s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrSemesterNotExist{id, ""}
		//return nil,nil
	}
	return s, nil
}

// GetSemesterByID returns the user object by given ID if exists.
func GetSemesterByID(id int64) (*Semester, error) {
	return getSemesterByID(x, id)
}

// GetSemesterByName returns user by given name.
func GetSemesterByName(name string) (*Semester, error) {
	if len(name) == 0 {
		return nil, ErrSemesterNotExist{0, name}
		//return nil,nil
	}
	s := &Semester{Name: name}
	has, err := x.Get(s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrSemesterNotExist{0, name}
		//return nil,nil
	}
	return s, nil
}

// GetSemesterIDsByNames returns a slice of ids corresponds to names.
func GetSemesterIDsByNames(names []string) []int64 {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		s, err := GetSemesterByName(name)
		if err != nil {
			continue
		}
		ids = append(ids, s.ID)
	}
	return ids
}

type SearchSemesterOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int // Can be smaller than or equal to setting.UI.ExplorePagingNum
}

// SearchSemesterByName takes keyword and part of user name to search,
// it returns results in given range and number of total results.
func SearchSemesterByName(opts *SearchSemesterOptions) (semesters []*Semester, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return semesters, 0, nil
	}
	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > setting.UI.ExplorePagingNum {
		opts.PageSize = setting.UI.ExplorePagingNum
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	semesters = make([]*Semester, 0, opts.PageSize)
	// Append conditions
	sess := x.Where("name LIKE ?", searchQuery)

	var countSess xorm.Session
	countSess = *sess
	count, err := countSess.Count(new(Semester))
	if err != nil {
		return nil, 0, fmt.Errorf("Count: %v", err)
	}

	if len(opts.OrderBy) > 0 {
		sess.OrderBy(opts.OrderBy)
	}
	return semesters, count, sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize).Find(&semesters)
}

func getSemesters() ([]*Semester, error) {
	semesters := make([]*Semester, 0, 5)
	return semesters, x.Asc("name").Find(&semesters)
}

func GetSemesters() ([]*Semester, error) {
	return getSemesters()
}

func GetSemestersProfessor(ProfessorID int64) ([]*Semester, error) {
	semesters := make([]*Semester, 0, 5)
	return semesters, x.Cols("semester.id", "semester.name").Join("LEFT", "`course`", "`course`.semester_id=`semester`.id").Where("course.uid=?", ProfessorID).Asc("name").GroupBy("semester.id").Find(&semesters)
}
