// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Unknwon/com"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
)

const (
	SETTINGS_PROFILE      base.TplName = "user/settings/profile"
	SETTINGS_AVATAR       base.TplName = "user/settings/avatar"
	SETTINGS_PASSWORD     base.TplName = "user/settings/password"
	SETTINGS_EMAILS       base.TplName = "user/settings/email"
	SETTINGS_COURSES      base.TplName = "user/settings/course"
	SETTINGS_COURSES_NEW  base.TplName = "user/settings/newcourse"
	SETTINGS_SSH_KEYS     base.TplName = "user/settings/sshkeys"
	SETTINGS_SOCIAL       base.TplName = "user/settings/social"
	SETTINGS_APPLICATIONS base.TplName = "user/settings/applications"
	SETTINGS_DELETE       base.TplName = "user/settings/delete"
	NOTIFICATION          base.TplName = "user/notification"
	SECURITY              base.TplName = "user/security"
)

func Settings(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsProfile"] = true
	ctx.HTML(200, SETTINGS_PROFILE)
}

func handleUsernameChange(ctx *context.Context, newName string) {
	// Non-local users are not allowed to change their username.
	if len(newName) == 0 || !ctx.User.IsLocal() {
		return
	}

	// Check if user name has been changed
	if ctx.User.LowerName != strings.ToLower(newName) {
		if err := models.ChangeUserName(ctx.User, newName); err != nil {
			switch {
			case models.IsErrUserAlreadyExist(err):
				ctx.Flash.Error(ctx.Tr("newName_been_taken"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings")
			case models.IsErrEmailAlreadyUsed(err):
				ctx.Flash.Error(ctx.Tr("form.email_been_used"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings")
			case models.IsErrNameReserved(err):
				ctx.Flash.Error(ctx.Tr("user.newName_reserved"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings")
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Flash.Error(ctx.Tr("user.newName_pattern_not_allowed"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings")
			default:
				ctx.Handle(500, "ChangeUserName", err)
			}
			return
		}
		log.Trace("User name changed: %s -> %s", ctx.User.Name, newName)
	}

	// In case it's just a case change
	ctx.User.Name = newName
	ctx.User.LowerName = strings.ToLower(newName)
}

func SettingsPost(ctx *context.Context, form auth.UpdateProfileForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsProfile"] = true

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_PROFILE)
		return
	}

	handleUsernameChange(ctx, form.Name)
	if ctx.Written() {
		return
	}

	ctx.User.FullName = form.FullName
	ctx.User.Email = form.Email
	ctx.User.Website = form.Website
	ctx.User.Location = form.Location
	if err := models.UpdateUser(ctx.User); err != nil {
		ctx.Handle(500, "UpdateUser", err)
		return
	}

	log.Trace("User settings updated: %s", ctx.User.Name)
	ctx.Flash.Success(ctx.Tr("settings.update_profile_success"))
	ctx.Redirect(setting.AppSubUrl + "/user/settings")
}

// FIXME: limit size.
func UpdateAvatarSetting(ctx *context.Context, form auth.AvatarForm, ctxUser *models.User) error {
	ctxUser.UseCustomAvatar = form.Source == auth.AVATAR_LOCAL
	if len(form.Gravatar) > 0 {
		ctxUser.Avatar = base.EncodeMD5(form.Gravatar)
		ctxUser.AvatarEmail = form.Gravatar
	}

	if form.Avatar != nil {
		fr, err := form.Avatar.Open()
		if err != nil {
			return fmt.Errorf("Avatar.Open: %v", err)
		}
		defer fr.Close()

		data, err := ioutil.ReadAll(fr)
		if err != nil {
			return fmt.Errorf("ioutil.ReadAll: %v", err)
		}
		if !base.IsImageFile(data) {
			return errors.New(ctx.Tr("settings.uploaded_avatar_not_a_image"))
		}
		if err = ctxUser.UploadAvatar(data); err != nil {
			return fmt.Errorf("UploadAvatar: %v", err)
		}
	} else {
		// No avatar is uploaded but setting has been changed to enable,
		// generate a random one when needed.
		if ctxUser.UseCustomAvatar && !com.IsFile(ctxUser.CustomAvatarPath()) {
			if err := ctxUser.GenerateRandomAvatar(); err != nil {
				log.Error(4, "GenerateRandomAvatar[%d]: %v", ctxUser.ID, err)
			}
		}
	}

	if err := models.UpdateUser(ctxUser); err != nil {
		return fmt.Errorf("UpdateUser: %v", err)
	}

	return nil
}

func SettingsAvatar(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsAvatar"] = true
	ctx.HTML(200, SETTINGS_AVATAR)
}

func SettingsAvatarPost(ctx *context.Context, form auth.AvatarForm) {
	if err := UpdateAvatarSetting(ctx, form, ctx.User); err != nil {
		ctx.Flash.Error(err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("settings.update_avatar_success"))
	}

	ctx.Redirect(setting.AppSubUrl + "/user/settings/avatar")
}

func SettingsDeleteAvatar(ctx *context.Context) {
	if err := ctx.User.DeleteAvatar(); err != nil {
		ctx.Flash.Error(err.Error())
	}

	ctx.Redirect(setting.AppSubUrl + "/user/settings/avatar")
}

func SettingsPassword(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsPassword"] = true
	ctx.HTML(200, SETTINGS_PASSWORD)
}

func SettingsPasswordPost(ctx *context.Context, form auth.ChangePasswordForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsPassword"] = true

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_PASSWORD)
		return
	}

	if !ctx.User.ValidatePassword(form.OldPassword) {
		ctx.Flash.Error(ctx.Tr("settings.password_incorrect"))
	} else if form.Password != form.Retype {
		ctx.Flash.Error(ctx.Tr("form.password_not_match"))
	} else {
		ctx.User.Passwd = form.Password
		ctx.User.Salt = models.GetUserSalt()
		ctx.User.EncodePasswd()
		if err := models.UpdateUser(ctx.User); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
		log.Trace("User password updated: %s", ctx.User.Name)
		ctx.Flash.Success(ctx.Tr("settings.change_password_success"))
	}

	ctx.Redirect(setting.AppSubUrl + "/user/settings/password")
}

func SettingsEmails(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsEmails"] = true

	emails, err := models.GetEmailAddresses(ctx.User.ID)
	if err != nil {
		ctx.Handle(500, "GetEmailAddresses", err)
		return
	}
	ctx.Data["Emails"] = emails

	ctx.HTML(200, SETTINGS_EMAILS)
}

func SettingsEmailPost(ctx *context.Context, form auth.AddEmailForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsEmails"] = true

	// Make emailaddress primary.
	if ctx.Query("_method") == "PRIMARY" {
		if err := models.MakeEmailPrimary(&models.EmailAddress{ID: ctx.QueryInt64("id")}); err != nil {
			ctx.Handle(500, "MakeEmailPrimary", err)
			return
		}

		log.Trace("Email made primary: %s", ctx.User.Name)
		ctx.Redirect(setting.AppSubUrl + "/user/settings/email")
		return
	}

	// Add Email address.
	emails, err := models.GetEmailAddresses(ctx.User.ID)
	if err != nil {
		ctx.Handle(500, "GetEmailAddresses", err)
		return
	}
	ctx.Data["Emails"] = emails

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_EMAILS)
		return
	}

	email := &models.EmailAddress{
		UID:         ctx.User.ID,
		Email:       form.Email,
		IsActivated: !setting.Service.RegisterEmailConfirm,
	}
	if err := models.AddEmailAddress(email); err != nil {
		if models.IsErrEmailAlreadyUsed(err) {
			ctx.RenderWithErr(ctx.Tr("form.email_been_used"), SETTINGS_EMAILS, &form)
			return
		}
		ctx.Handle(500, "AddEmailAddress", err)
		return
	}

	// Send confirmation email
	if setting.Service.RegisterEmailConfirm {
		models.SendActivateEmailMail(ctx.Context, ctx.User, email)

		if err := ctx.Cache.Put("MailResendLimit_"+ctx.User.LowerName, ctx.User.LowerName, 180); err != nil {
			log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
		}
		ctx.Flash.Info(ctx.Tr("settings.add_email_confirmation_sent", email.Email, setting.Service.ActiveCodeLives/60))
	} else {
		ctx.Flash.Success(ctx.Tr("settings.add_email_success"))
	}

	log.Trace("Email address added: %s", email.Email)
	ctx.Redirect(setting.AppSubUrl + "/user/settings/email")
}

func DeleteEmail(ctx *context.Context) {
	if err := models.DeleteEmailAddress(&models.EmailAddress{ID: ctx.QueryInt64("id")}); err != nil {
		ctx.Handle(500, "DeleteEmail", err)
		return
	}
	log.Trace("Email address deleted: %s", ctx.User.Name)

	ctx.Flash.Success(ctx.Tr("settings.email_deletion_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/user/settings/email",
	})
}

func SettingsCourses(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsSubjects"] = true

	courses, err := ctx.User.GetCoursesInfo()
	if err != nil {
		ctx.Handle(500, "GetCoursesInfo", err)
		return
	}
	ctx.Data["Courses"] = courses

	ctx.HTML(200, SETTINGS_COURSES)
}

func PrepareCoursesInfo(ctx *context.Context) {
	subjects, err := models.GetSubjects()
	if err != nil {
		ctx.Handle(500, "GetSubjects", err)
		return
	}
	ctx.Data["Subjects"] = subjects

	semesters, err := models.GetSemesters()
	if err != nil {
		ctx.Handle(500, "GetSemesters", err)
		return
	}
	ctx.Data["Semesters"] = semesters

	groups, err := models.GetGroups()
	if err != nil {
		ctx.Handle(500, "GetGroups", err)
		return
	}
	ctx.Data["Groups"] = groups
}

func NewCourse(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsSubjects"] = true

	PrepareCoursesInfo(ctx)

	ctx.HTML(200, SETTINGS_COURSES_NEW)
}

func NewCoursePost(ctx *context.Context, form auth.CreateNewCourseForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsSubjects"] = true

	semesterID := form.Semester
	groupID := form.Group
	subjectID := form.Subject
	estatus := form.Estatus

	if err := ctx.User.AddCourse(subjectID, semesterID, groupID, estatus); err != nil {

		switch {
		case models.IsErrCourseAlreadyExist(err):
			PrepareCoursesInfo(ctx)
			ctx.Data["Err_Course"] = true
			ctx.RenderWithErr(ctx.Tr("user.settings.course.already_exists"), SETTINGS_COURSES_NEW, &form)
		default:
			ctx.Handle(500, "AddCourse", err)
		}

		return
	}

	ctx.Flash.Success(ctx.Tr("user.settings.course.add_course_success"))
	ctx.Redirect(setting.AppSubUrl + "/user/settings/course")
}

func CoursePost(ctx *context.Context, form auth.AdminCrateSubjectForm) {
	/*name := ctx.Query("subject")
	if len(name) == 0 {
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
		return
	}

	s, err := models.GetSubjectByName(name)
	if err != nil {
		if models.IsErrSubjectNotExist(err) {

			subject := &models.Subject{
				Name: name,
			}
			if err := models.CreateSubject(subject); err != nil {
				switch {
				case models.IsErrSubjectAlreadyExist(err):
					ctx.Data["Err_SubjectName"] = true
					ctx.RenderWithErr(ctx.Tr("form.subjectname_been_taken"), SETTINGS_COURSES, &form)
				case models.IsErrNameReserved(err):
					ctx.Data["Err_SubjectName"] = true
					ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), SETTINGS_COURSES, &form)
				case models.IsErrNamePatternNotAllowed(err):
					ctx.Data["Err_SubjectName"] = true
					ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SETTINGS_COURSES, &form)
				default:
					ctx.Handle(500, "CreateSubject", err)
				}
				return
			}
			s, err = models.GetSubjectByName(name)
			log.Trace("Subject created by (%s): %s", ctx.User.Name, subject.Name)
		} else {
			ctx.Handle(500, "GetSubjectByName", err)
		}
	}

	// Check if user is organization member.
	if s.IsUserSubj(ctx.User.ID) {
		ctx.Flash.Error(ctx.Tr("user.settings.course.subject_is_user_course"))
		ctx.Redirect(setting.AppSubUrl + "/user/settings/course")
		return
	}

	if err = ctx.User.AddCourse(s.ID); err != nil {
		ctx.Handle(500, "AddCollaborator", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("user.settings.course.add_course_success"))
	ctx.Redirect(setting.AppSubUrl + "/user/settings/course") */
}

func ChangeCourseStatus(ctx *context.Context) {
	if err := ctx.User.ChangeCourseStatus(
		ctx.QueryInt64("sid"),
		ctx.QueryInt("status")); err != nil {
		log.Error(4, "ChangeCourseStatus: %v", err)
	}
}

func DeleteCourse(ctx *context.Context) {
	if err := ctx.User.RemoveCourse(ctx.QueryInt64("id")); err != nil {
		ctx.Handle(500, "RemoveCourse", err)
		return
	}
	log.Trace("Removed course for profesor: %s", ctx.User.Name)

	ctx.Flash.Success(ctx.Tr("user.settings.course.course_deletion_success"))
	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/user/settings/course",
	})
}

func SettingsSSHKeys(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsSSHKeys"] = true

	keys, err := models.ListPublicKeys(ctx.User.ID)
	if err != nil {
		ctx.Handle(500, "ListPublicKeys", err)
		return
	}
	ctx.Data["Keys"] = keys

	ctx.HTML(200, SETTINGS_SSH_KEYS)
}

func SettingsSSHKeysPost(ctx *context.Context, form auth.AddSSHKeyForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsSSHKeys"] = true

	keys, err := models.ListPublicKeys(ctx.User.ID)
	if err != nil {
		ctx.Handle(500, "ListPublicKeys", err)
		return
	}
	ctx.Data["Keys"] = keys

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_SSH_KEYS)
		return
	}

	content, err := models.CheckPublicKeyString(form.Content)
	if err != nil {
		if models.IsErrKeyUnableVerify(err) {
			ctx.Flash.Info(ctx.Tr("form.unable_verify_ssh_key"))
		} else {
			ctx.Flash.Error(ctx.Tr("form.invalid_ssh_key", err.Error()))
			ctx.Redirect(setting.AppSubUrl + "/user/settings/ssh")
			return
		}
	}

	if _, err = models.AddPublicKey(ctx.User.ID, form.Title, content); err != nil {
		ctx.Data["HasError"] = true
		switch {
		case models.IsErrKeyAlreadyExist(err):
			ctx.Data["Err_Content"] = true
			ctx.RenderWithErr(ctx.Tr("settings.ssh_key_been_used"), SETTINGS_SSH_KEYS, &form)
		case models.IsErrKeyNameAlreadyUsed(err):
			ctx.Data["Err_Title"] = true
			ctx.RenderWithErr(ctx.Tr("settings.ssh_key_name_used"), SETTINGS_SSH_KEYS, &form)
		default:
			ctx.Handle(500, "AddPublicKey", err)
		}
		return
	}

	ctx.Flash.Success(ctx.Tr("settings.add_key_success", form.Title))
	ctx.Redirect(setting.AppSubUrl + "/user/settings/ssh")
}

func DeleteSSHKey(ctx *context.Context) {
	if err := models.DeletePublicKey(ctx.User, ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeletePublicKey: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("settings.ssh_key_deletion_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/user/settings/ssh",
	})
}

func SettingsApplications(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsApplications"] = true

	tokens, err := models.ListAccessTokens(ctx.User.ID)
	if err != nil {
		ctx.Handle(500, "ListAccessTokens", err)
		return
	}
	ctx.Data["Tokens"] = tokens

	ctx.HTML(200, SETTINGS_APPLICATIONS)
}

func SettingsApplicationsPost(ctx *context.Context, form auth.NewAccessTokenForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsApplications"] = true

	if ctx.HasError() {
		tokens, err := models.ListAccessTokens(ctx.User.ID)
		if err != nil {
			ctx.Handle(500, "ListAccessTokens", err)
			return
		}
		ctx.Data["Tokens"] = tokens
		ctx.HTML(200, SETTINGS_APPLICATIONS)
		return
	}

	t := &models.AccessToken{
		UID:  ctx.User.ID,
		Name: form.Name,
	}
	if err := models.NewAccessToken(t); err != nil {
		ctx.Handle(500, "NewAccessToken", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("settings.generate_token_succees"))
	ctx.Flash.Info(t.Sha1)

	ctx.Redirect(setting.AppSubUrl + "/user/settings/applications")
}

func SettingsDeleteApplication(ctx *context.Context) {
	if err := models.DeleteAccessTokenByID(ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeleteAccessTokenByID: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("settings.delete_token_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubUrl + "/user/settings/applications",
	})
}

func SettingsDelete(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsDelete"] = true

	if ctx.Req.Method == "POST" {
		if _, err := models.UserSignIn(ctx.User.Name, ctx.Query("password")); err != nil {
			if models.IsErrUserNotExist(err) {
				ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_password"), SETTINGS_DELETE, nil)
			} else {
				ctx.Handle(500, "UserSignIn", err)
			}
			return
		}

		if err := models.DeleteUser(ctx.User); err != nil {
			switch {
			case models.IsErrUserOwnRepos(err):
				ctx.Flash.Error(ctx.Tr("form.still_own_repo"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings/delete")
			case models.IsErrUserHasOrgs(err):
				ctx.Flash.Error(ctx.Tr("form.still_has_org"))
				ctx.Redirect(setting.AppSubUrl + "/user/settings/delete")
			default:
				ctx.Handle(500, "DeleteUser", err)
			}
		} else {
			log.Trace("Account deleted: %s", ctx.User.Name)
			ctx.Redirect(setting.AppSubUrl + "/")
		}
		return
	}

	ctx.HTML(200, SETTINGS_DELETE)
}
