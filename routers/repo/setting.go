// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"github.com/gogits/git-module"
	"strconv"
	"strings"
	"time"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
)

const (
	SETTINGS_OPTIONS base.TplName = "repo/settings/options"
	COLLABORATION    base.TplName = "repo/settings/collaboration"
	TAGS             base.TplName = "repo/settings/tags_manage"
	SCHOOL           base.TplName = "repo/settings/school_data"
	GITHOOKS         base.TplName = "repo/settings/githooks"
	GITHOOK_EDIT     base.TplName = "repo/settings/githook_edit"
	DEPLOY_KEYS      base.TplName = "repo/settings/deploy_keys"
)

func Settings(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings")
	ctx.Data["PageIsSettingsOptions"] = true

	professors, err := models.GetProfessors()
	if err != nil {
		ctx.Handle(500, "GetProfessors", err)
		return
	}

	ctx.Data["Professors"] = professors

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

	repo := ctx.Repo.Repository
	if repo.ProfessorID > 0 {
		professor, err := repo.GetProfessor()
		if err != nil {
			ctx.Handle(500, "GetProfessor", err)
			return
		}
		ctx.Data["Professor"] = professor
		ctx.Data["Professor_Success"] = true
	}

	if repo.SubjectID > 0 {
		subject, err := repo.GetSubject()
		if err != nil {
			ctx.Handle(500, "GetSubject", err)
			return
		}
		ctx.Data["Subject"] = subject
		ctx.Data["Subject_Success"] = true
	}

	if repo.SemesterID > 0 {
		semester, err := repo.GetSemester()
		if err != nil {
			ctx.Handle(500, "GetSemester", err)
			return
		}
		ctx.Data["Semester"] = semester
		ctx.Data["Semester_Success"] = true
	}

	if repo.GroupID > 0 {
		group, err := repo.GetGroup()
		if err != nil {
			ctx.Handle(500, "GetGroup", err)
			return
		}
		ctx.Data["Group"] = group
		ctx.Data["Group_Success"] = true
	}

	ctx.HTML(200, SETTINGS_OPTIONS)
}

func SettingsPost(ctx *context.Context, form auth.RepoSettingForm) {
	ctx.Data["Title"] = ctx.Tr("repo.settings")
	ctx.Data["PageIsSettingsOptions"] = true

	repo := ctx.Repo.Repository

	switch ctx.Query("action") {
	case "update":
		if ctx.HasError() {
			ctx.HTML(200, SETTINGS_OPTIONS)
			return
		}

		isNameChanged := false
		oldRepoName := repo.Name
		newRepoName := form.RepoName
		// Check if repository name has been changed.
		if repo.LowerName != strings.ToLower(newRepoName) {
			isNameChanged = true
			if err := models.ChangeRepositoryName(ctx.Repo.Owner, repo.Name, newRepoName); err != nil {
				ctx.Data["Err_RepoName"] = true
				switch {
				case models.IsErrRepoAlreadyExist(err):
					ctx.RenderWithErr(ctx.Tr("form.repo_name_been_taken"), SETTINGS_OPTIONS, &form)
				case models.IsErrNameReserved(err):
					ctx.RenderWithErr(ctx.Tr("repo.form.name_reserved", err.(models.ErrNameReserved).Name), SETTINGS_OPTIONS, &form)
				case models.IsErrNamePatternNotAllowed(err):
					ctx.RenderWithErr(ctx.Tr("repo.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SETTINGS_OPTIONS, &form)
				default:
					ctx.Handle(500, "ChangeRepositoryName", err)
				}
				return
			}

			log.Trace("Repository name changed: %s/%s -> %s", ctx.Repo.Owner.Name, repo.Name, newRepoName)
		}
		// In case it's just a case change.
		repo.Name = newRepoName
		repo.LowerName = strings.ToLower(newRepoName)

		if ctx.Repo.GitRepo.IsBranchExist(form.Branch) &&
			repo.DefaultBranch != form.Branch {
			repo.DefaultBranch = form.Branch
			if err := ctx.Repo.GitRepo.SetDefaultBranch(form.Branch); err != nil {
				if !git.IsErrUnsupportedVersion(err) {
					ctx.Handle(500, "SetDefaultBranch", err)
					return
				}
			}
		}
		repo.Description = form.Description
		repo.Website = form.Website

		//METADATOS ESCOLARES
		if repo.ProfessorID != form.Professor {
			if err := repo.DeleteCollaboration(repo.ProfessorID); err != nil {
				log.Error(4, "DeleteCollaboration: ", err)
			}

			repo.ProfessorID = form.Professor

			u, err := models.GetUserByID(ctx.QueryInt64("professor"))
			if err != nil {
				if models.IsErrUserNotExist(err) {
					log.Error(4, "IsErrUserNotExist: ", err)
				}
				return
			}

			if err := repo.AddCollaborator(u); err != nil {
				log.Error(4, "AddCollaborator: ", err)
				return
			}
		}

		repo.SubjectID = form.Subject
		repo.SemesterID = form.Semester
		repo.GroupID = form.Group
		// FIN METADATOS ESCOLARES

		// Visibility of forked repository is forced sync with base repository.
		if repo.IsFork {
			form.Private = repo.BaseRepo.IsPrivate
		}

		visibilityChanged := repo.IsPrivate != form.Private
		repo.IsPrivate = form.Private
		if err := models.UpdateRepository(repo, visibilityChanged); err != nil {
			ctx.Handle(500, "UpdateRepository", err)
			return
		}
		log.Trace("Repository basic settings updated: %s/%s", ctx.Repo.Owner.Name, repo.Name)

		if isNameChanged {
			if err := models.RenameRepoAction(ctx.User, oldRepoName, repo); err != nil {
				log.Error(4, "RenameRepoAction: %v", err)
			}
		}

		ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
		ctx.Redirect(repo.Link() + "/settings")

	case "mirror":
		if !repo.IsMirror {
			ctx.Handle(404, "", nil)
			return
		}

		if form.Interval > 0 {
			ctx.Repo.Mirror.EnablePrune = form.EnablePrune
			ctx.Repo.Mirror.Interval = form.Interval
			ctx.Repo.Mirror.NextUpdate = time.Now().Add(time.Duration(form.Interval) * time.Hour)
			if err := models.UpdateMirror(ctx.Repo.Mirror); err != nil {
				ctx.Handle(500, "UpdateMirror", err)
				return
			}
		}
		if err := ctx.Repo.Mirror.SaveAddress(form.MirrorAddress); err != nil {
			ctx.Handle(500, "SaveAddress", err)
			return
		}

		ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
		ctx.Redirect(repo.Link() + "/settings")

	case "mirror-sync":
		if !repo.IsMirror {
			ctx.Handle(404, "", nil)
			return
		}

		go models.MirrorQueue.Add(repo.ID)
		ctx.Flash.Info(ctx.Tr("repo.settings.mirror_sync_in_progress"))
		ctx.Redirect(repo.Link() + "/settings")

	case "advanced":
		repo.EnableWiki = form.EnableWiki
		repo.EnableExternalWiki = form.EnableExternalWiki
		repo.ExternalWikiURL = form.ExternalWikiURL
		repo.EnableIssues = form.EnableIssues
		repo.EnableExternalTracker = form.EnableExternalTracker
		repo.ExternalTrackerFormat = form.TrackerURLFormat
		repo.ExternalTrackerStyle = form.TrackerIssueStyle
		repo.EnablePulls = form.EnablePulls

		if err := models.UpdateRepository(repo, false); err != nil {
			ctx.Handle(500, "UpdateRepository", err)
			return
		}
		log.Trace("Repository advanced settings updated: %s/%s", ctx.Repo.Owner.Name, repo.Name)

		ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
		ctx.Redirect(ctx.Repo.RepoLink + "/settings")

	case "convert":
		if !ctx.Repo.IsOwner() {
			ctx.Error(404)
			return
		}
		if repo.Name != form.RepoName {
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_repo_name"), SETTINGS_OPTIONS, nil)
			return
		}

		if ctx.Repo.Owner.IsOrganization() {
			if !ctx.Repo.Owner.IsOwnedBy(ctx.User.ID) {
				ctx.Error(404)
				return
			}
		}

		if !repo.IsMirror {
			ctx.Error(404)
			return
		}
		repo.IsMirror = false

		if _, err := models.CleanUpMigrateInfo(repo); err != nil {
			ctx.Handle(500, "CleanUpMigrateInfo", err)
			return
		} else if err = models.DeleteMirrorByRepoID(ctx.Repo.Repository.ID); err != nil {
			ctx.Handle(500, "DeleteMirrorByRepoID", err)
			return
		}
		log.Trace("Repository converted from mirror to regular: %s/%s", ctx.Repo.Owner.Name, repo.Name)
		ctx.Flash.Success(ctx.Tr("repo.settings.convert_succeed"))
		ctx.Redirect(setting.AppSubUrl + "/" + ctx.Repo.Owner.Name + "/" + repo.Name)

	case "transfer":
		if !ctx.Repo.IsOwner() {
			ctx.Error(404)
			return
		}
		if repo.Name != form.RepoName {
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_repo_name"), SETTINGS_OPTIONS, nil)
			return
		}

		if ctx.Repo.Owner.IsOrganization() {
			if !ctx.Repo.Owner.IsOwnedBy(ctx.User.ID) {
				ctx.Error(404)
				return
			}
		}

		newOwner := ctx.Query("new_owner_name")
		isExist, err := models.IsUserExist(0, newOwner)
		if err != nil {
			ctx.Handle(500, "IsUserExist", err)
			return
		} else if !isExist {
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_owner_name"), SETTINGS_OPTIONS, nil)
			return
		}

		if err = models.TransferOwnership(ctx.User, newOwner, repo); err != nil {
			if models.IsErrRepoAlreadyExist(err) {
				ctx.RenderWithErr(ctx.Tr("repo.settings.new_owner_has_same_repo"), SETTINGS_OPTIONS, nil)
			} else {
				ctx.Handle(500, "TransferOwnership", err)
			}
			return
		}
		log.Trace("Repository transfered: %s/%s -> %s", ctx.Repo.Owner.Name, repo.Name, newOwner)
		ctx.Flash.Success(ctx.Tr("repo.settings.transfer_succeed"))
		ctx.Redirect(setting.AppSubUrl + "/" + newOwner + "/" + repo.Name)

	case "delete":
		if !ctx.Repo.IsOwner() {
			ctx.Error(404)
			return
		}
		if repo.Name != form.RepoName {
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_repo_name"), SETTINGS_OPTIONS, nil)
			return
		}

		if ctx.Repo.Owner.IsOrganization() {
			if !ctx.Repo.Owner.IsOwnedBy(ctx.User.ID) {
				ctx.Error(404)
				return
			}
		}

		if err := models.DeleteRepository(ctx.Repo.Owner.ID, repo.ID); err != nil {
			ctx.Handle(500, "DeleteRepository", err)
			return
		}
		log.Trace("Repository deleted: %s/%s", ctx.Repo.Owner.Name, repo.Name)

		ctx.Flash.Success(ctx.Tr("repo.settings.deletion_success"))
		ctx.Redirect(ctx.Repo.Owner.DashboardLink())

	case "delete-wiki":
		if !ctx.Repo.IsOwner() {
			ctx.Error(404)
			return
		}
		if repo.Name != form.RepoName {
			ctx.RenderWithErr(ctx.Tr("form.enterred_invalid_repo_name"), SETTINGS_OPTIONS, nil)
			return
		}

		if ctx.Repo.Owner.IsOrganization() {
			if !ctx.Repo.Owner.IsOwnedBy(ctx.User.ID) {
				ctx.Error(404)
				return
			}
		}

		repo.DeleteWiki()
		log.Trace("Repository wiki deleted: %s/%s", ctx.Repo.Owner.Name, repo.Name)

		repo.EnableWiki = false
		if err := models.UpdateRepository(repo, false); err != nil {
			ctx.Handle(500, "UpdateRepository", err)
			return
		}

		ctx.Flash.Success(ctx.Tr("repo.settings.wiki_deletion_success"))
		ctx.Redirect(ctx.Repo.RepoLink + "/settings")

	default:
		ctx.Handle(404, "", nil)
	}
}

func ManageTag(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.tags")
	ctx.Data["PageIsSettingsTags"] = true

	repoID := ctx.Repo.Repository.ID
	tags, err := models.GetTagsOfRepo(repoID)

	if err != nil {
		ctx.Handle(500, "GetTagsOfRepo", err)
		return
	}
	ctx.Data["Tags"] = tags

	tags2, err := models.GetTags()
	if err != nil {
		ctx.Handle(500, "GetTags", err)
		return
	}

	ctx.Data["TagsR"] = tags2

	ctx.HTML(200, TAGS)
}

func ManageTagPost(ctx *context.Context) {

	switch ctx.Query("action") {
	case "update":
		tags := strings.ToLower(ctx.Query("tags"))
		repoID := ctx.Repo.Repository.ID

		arr_tags := strings.Split(tags, ",")
		for _, tag := range arr_tags {
			tagID, _ := strconv.ParseInt(tag, 10, 64)
			models.LinkTagtoRepo(tagID, repoID, true)
		}

		ctx.Flash.Success(ctx.Tr("repo.settings.add_tag_success"))
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)

	case "create":

		t := &models.Tag{
			Etiqueta: ctx.Query("etiqueta"),
		}

		if err := models.CreateTag(t); err != nil {
			switch {
			case models.IsErrTagAlreadyExist(err):
				ctx.Data["Err_TagName"] = true
				ctx.Flash.Error(ctx.Tr("form.tagname_been_taken"))
				ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
			case models.IsErrNameReserved(err):
				ctx.Data["Err_TagName"] = true
				ctx.Flash.Error(ctx.Tr("user.form.name_reserved"))
				ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
			case models.IsErrNamePatternNotAllowed(err):
				ctx.Data["Err_TagName"] = true
				ctx.Flash.Error(ctx.Tr("user.form.name_pattern_not_allowed"))
				ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
			default:
				ctx.Handle(500, "Create Tag User", err)
			}
			return
		}

		log.Trace("Tag created: %s", t.Etiqueta)

		ctx.Flash.Success(ctx.Tr("admin.new_tag_success", t.Etiqueta))
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)

	default:
		ctx.Handle(404, "", nil)
	}
}

func DeleteTagsRepo(ctx *context.Context) {
	if err := ctx.Repo.Repository.DeleteCollaboration(ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeleteCollaboration: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.remove_tag_success"))
	}

	repoID := ctx.Repo.Repository.ID
	borrar := models.UnlinkTagRepo(repoID, ctx.QueryInt64("id"))

	if borrar {
		ctx.Flash.Success(ctx.Tr("repo.settings.remove_tag_success"))
	} else {
		ctx.Flash.Error("Error unlinking tag")
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": ctx.Repo.RepoLink + "/settings/tags",
	})
}

func Collaboration(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings")
	ctx.Data["PageIsSettingsCollaboration"] = true

	users, err := ctx.Repo.Repository.GetCollaborators()
	if err != nil {
		ctx.Handle(500, "GetCollaborators", err)
		return
	}
	ctx.Data["Collaborators"] = users

	ctx.HTML(200, COLLABORATION)
}

func CollaborationPost(ctx *context.Context) {
	name := strings.ToLower(ctx.Query("collaborator"))
	if len(name) == 0 || ctx.Repo.Owner.LowerName == name {
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
		return
	}

	if strings.Contains(name, "@"){
		models.SendRegisterInvitationCollab(name, ctx.User, ctx.Repo.Repository)
		ctx.Flash.Success(ctx.Tr("repo.settings.add_collaborator_reg_success", name))
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
	} else {

		u, err := models.GetUserByName(name)
		if err != nil {
			if models.IsErrUserNotExist(err) {
				ctx.Flash.Error(ctx.Tr("form.user_not_exist"))
				ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
			} else {
				ctx.Handle(500, "GetUserByName", err)
			}
			return
		}

		// Organization is not allowed to be added as a collaborator.
		if u.IsOrganization() {
			ctx.Flash.Error(ctx.Tr("repo.settings.org_not_allowed_to_be_collaborator"))
			ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
			return
		}

		// Check if user is organization member.
		if ctx.Repo.Owner.IsOrganization() && ctx.Repo.Owner.IsOrgMember(u.ID) {
			ctx.Flash.Info(ctx.Tr("repo.settings.user_is_org_member"))
			ctx.Redirect(ctx.Repo.RepoLink + "/settings/collaboration")
			return
		}

		if err = ctx.Repo.Repository.AddCollaborator(u); err != nil {
			ctx.Handle(500, "AddCollaborator", err)
			return
		}

		if setting.Service.EnableNotifyMail {
			models.SendCollaboratorMail(u, ctx.User, ctx.Repo.Repository)
		}

		ctx.Flash.Success(ctx.Tr("repo.settings.add_collaborator_success"))
		ctx.Redirect(setting.AppSubUrl + ctx.Req.URL.Path)
	}
}

func ChangeCollaborationAccessMode(ctx *context.Context) {
	if err := ctx.Repo.Repository.ChangeCollaborationAccessMode(
		ctx.QueryInt64("uid"),
		models.AccessMode(ctx.QueryInt("mode"))); err != nil {
		log.Error(4, "ChangeCollaborationAccessMode: %v", err)
	}
}

func DeleteCollaboration(ctx *context.Context) {
	if err := ctx.Repo.Repository.DeleteCollaboration(ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeleteCollaboration: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.remove_collaborator_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": ctx.Repo.RepoLink + "/settings/collaboration",
	})
}

func parseOwnerAndRepo(ctx *context.Context) (*models.User, *models.Repository) {
	owner, err := models.GetUserByName(ctx.Params(":username"))
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.Handle(404, "GetUserByName", err)
		} else {
			ctx.Handle(500, "GetUserByName", err)
		}
		return nil, nil
	}

	repo, err := models.GetRepositoryByName(owner.ID, ctx.Params(":reponame"))
	if err != nil {
		if models.IsErrRepoNotExist(err) {
			ctx.Handle(404, "GetRepositoryByName", err)
		} else {
			ctx.Handle(500, "GetRepositoryByName", err)
		}
		return nil, nil
	}

	return owner, repo
}

func GitHooks(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.githooks")
	ctx.Data["PageIsSettingsGitHooks"] = true

	hooks, err := ctx.Repo.GitRepo.Hooks()
	if err != nil {
		ctx.Handle(500, "Hooks", err)
		return
	}
	ctx.Data["Hooks"] = hooks

	ctx.HTML(200, GITHOOKS)
}

func GitHooksEdit(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.githooks")
	ctx.Data["PageIsSettingsGitHooks"] = true

	name := ctx.Params(":name")
	hook, err := ctx.Repo.GitRepo.GetHook(name)
	if err != nil {
		if err == git.ErrNotValidHook {
			ctx.Handle(404, "GetHook", err)
		} else {
			ctx.Handle(500, "GetHook", err)
		}
		return
	}
	ctx.Data["Hook"] = hook
	ctx.HTML(200, GITHOOK_EDIT)
}

func GitHooksEditPost(ctx *context.Context) {
	name := ctx.Params(":name")
	hook, err := ctx.Repo.GitRepo.GetHook(name)
	if err != nil {
		if err == git.ErrNotValidHook {
			ctx.Handle(404, "GetHook", err)
		} else {
			ctx.Handle(500, "GetHook", err)
		}
		return
	}
	hook.Content = ctx.Query("content")
	if err = hook.Update(); err != nil {
		ctx.Handle(500, "hook.Update", err)
		return
	}
	ctx.Redirect(ctx.Repo.RepoLink + "/settings/hooks/git")
}

func DeployKeys(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.deploy_keys")
	ctx.Data["PageIsSettingsKeys"] = true

	keys, err := models.ListDeployKeys(ctx.Repo.Repository.ID)
	if err != nil {
		ctx.Handle(500, "ListDeployKeys", err)
		return
	}
	ctx.Data["Deploykeys"] = keys

	ctx.HTML(200, DEPLOY_KEYS)
}

func DeployKeysPost(ctx *context.Context, form auth.AddSSHKeyForm) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.deploy_keys")
	ctx.Data["PageIsSettingsKeys"] = true

	keys, err := models.ListDeployKeys(ctx.Repo.Repository.ID)
	if err != nil {
		ctx.Handle(500, "ListDeployKeys", err)
		return
	}
	ctx.Data["Deploykeys"] = keys

	if ctx.HasError() {
		ctx.HTML(200, DEPLOY_KEYS)
		return
	}

	content, err := models.CheckPublicKeyString(form.Content)
	if err != nil {
		if models.IsErrKeyUnableVerify(err) {
			ctx.Flash.Info(ctx.Tr("form.unable_verify_ssh_key"))
		} else {
			ctx.Data["HasError"] = true
			ctx.Data["Err_Content"] = true
			ctx.Flash.Error(ctx.Tr("form.invalid_ssh_key", err.Error()))
			ctx.Redirect(ctx.Repo.RepoLink + "/settings/keys")
			return
		}
	}

	key, err := models.AddDeployKey(ctx.Repo.Repository.ID, form.Title, content)
	if err != nil {
		ctx.Data["HasError"] = true
		switch {
		case models.IsErrKeyAlreadyExist(err):
			ctx.Data["Err_Content"] = true
			ctx.RenderWithErr(ctx.Tr("repo.settings.key_been_used"), DEPLOY_KEYS, &form)
		case models.IsErrKeyNameAlreadyUsed(err):
			ctx.Data["Err_Title"] = true
			ctx.RenderWithErr(ctx.Tr("repo.settings.key_name_used"), DEPLOY_KEYS, &form)
		default:
			ctx.Handle(500, "AddDeployKey", err)
		}
		return
	}

	log.Trace("Deploy key added: %d", ctx.Repo.Repository.ID)
	ctx.Flash.Success(ctx.Tr("repo.settings.add_key_success", key.Name))
	ctx.Redirect(ctx.Repo.RepoLink + "/settings/keys")
}

func DeleteDeployKey(ctx *context.Context) {
	if err := models.DeleteDeployKey(ctx.User, ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeleteDeployKey: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.deploy_key_deletion_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": ctx.Repo.RepoLink + "/settings/keys",
	})
}
