// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
)


// Course represents an subject-user relation.
type Course struct {
	ID       int64 `xorm:"pk autoincr"`
	Uid      int64 `xorm:"INDEX UNIQUE(s)"`
	SubjID   int64 `xorm:"INDEX UNIQUE(s)"`
	IsActive bool
}

func (s *Subject) IsUserSubj(uid int64) bool {
	return IsUserSubject(s.ID, uid)
}

func IsUserSubject(subjID, uid int64) bool {
	has, _ := x.Where("uid=?", uid).And("subj_id=?", subjID).Get(new(Course))
	return has
}

// AddCourse adds new course.
func (u *User) AddCourse(subjID int64) error {
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	c := &Course{
		Uid:   		u.ID,
		SubjID: 	subjID,
		IsActive: 	true,
	}

	if _, err := sess.Insert(c); err != nil {
		sess.Rollback()
		return err
	}

	return sess.Commit()
}

func (u *User) getCourses(e Engine) ([]*Course, error) {
	courses := make([]*Course, 0)
	return courses, e.Find(&courses, &Course{Uid: u.ID})
}

type CourseInfo struct {
	*Subject
	Course *Course
}

func (u *User) GetCoursesInfo() ([]*CourseInfo, error) {
	return u.getCoursesInfo(x)
}

func (u *User) getCoursesInfo(e Engine) ([]*CourseInfo, error) {
	courses, err := u.getCourses(e)
	if err != nil {
		return nil, fmt.Errorf("getCourses: %v", err)
	}

	info := make([]*CourseInfo, len(courses))
	for i, c := range courses {
		subject, err := GetSubjectByID(c.SubjID)
		if err != nil {
			return nil, err
		}
		info[i] = &CourseInfo{
			Subject: subject,
			Course:  c,
		}
	}
	return info, nil
}

// ChangeCourseStatus changes active or inactive status.
func (u *User) ChangeCourseStatus(subjID int64, active int) error {
	c := new(Course)
	has, err := x.Where("uid=?", u.ID).And("subj_id=?", subjID).Get(c)
	if err != nil {
		return err
	} else if !has {
		return nil
	}

	status := false

	if active == 1 {
		status = true
	} else {
		status = false
	}

	c.IsActive = status 
	_, err = x.Id(c.ID).AllCols().Update(c)
	return err
}


// RemoveCourse removes user from given subject.
func (u *User) RemoveCourse(subjID int64) error {
	c := new(Course)

	has, err := x.Where("uid=?", u.ID).And("subj_id=?", subjID).Get(c)
	if err != nil {
		return fmt.Errorf("get course: %v", err)
	} else if !has {
		return nil
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(c.ID).Delete(c); err != nil {
		return err
	}

	return sess.Commit()
}
