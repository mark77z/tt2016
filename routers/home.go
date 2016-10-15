// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"fmt"

	"github.com/Unknwon/paginater"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
	"github.com/gogits/gogs/routers/user"
)

const (
	HOME                  base.TplName = "home"
	EXPLORE_REPOS         base.TplName = "explore/repos"
	EXPLORE_ADV           base.TplName = "explore/advanced"
	EXPLORE_USERS         base.TplName = "explore/users"
	EXPLORE_ORGANIZATIONS base.TplName = "explore/organizations"
	EXPLORE_SUBJECTS      base.TplName = "explore/subjects"
	EXPLORE_SEMESTERS     base.TplName = "explore/semesters"
	EXPLORE_GROUPS        base.TplName = "explore/groups"
	EXPLORE_TAGS          base.TplName = "explore/tags"
	EXPLORE_PROFESSORS    base.TplName = "explore/professors"
)

func Home(ctx *context.Context) {
	if ctx.IsSigned {
		if !ctx.User.IsActive && setting.Service.RegisterEmailConfirm {
			ctx.Data["Title"] = ctx.Tr("auth.active_your_account")
			ctx.HTML(200, user.ACTIVATE)
		} else {
			user.Dashboard(ctx)
		}
		return
	}

	// Check auto-login.
	uname := ctx.GetCookie(setting.CookieUserName)
	if len(uname) != 0 {
		ctx.Redirect(setting.AppSubUrl + "/user/login")
		return
	}

	ctx.Data["PageIsHome"] = true
	ctx.HTML(200, HOME)
}

type RepoSearchOptions struct {
	Counter  func(bool) int64
	Ranger   func(int, int) ([]*models.Repository, error)
	Private  bool
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderRepoSearch(ctx *context.Context, opts *RepoSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		repos []*models.Repository
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		repos, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter(opts.Private)
	} else {
		repos, count, err = models.SearchRepositoryByName(&models.SearchRepoOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Private:  opts.Private,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchRepositoryByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)

	for _, repo := range repos {
		if err = repo.GetOwner(); err != nil {
			log.Trace("%s", repo)
			ctx.Handle(500, "GetOwner", fmt.Errorf("%d: %v", repo.ID, err))
			return
		}
	}
	ctx.Data["Repos"] = repos

	ctx.HTML(200, opts.TplName)
}

func ExploreRepos(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreRepositories"] = true

	RenderRepoSearch(ctx, &RepoSearchOptions{
		Counter:  models.CountRepositories,
		Ranger:   models.GetRecentUpdatedRepositories,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "repository.updated_unix DESC",
		TplName:  EXPLORE_REPOS,
	})
}

func RenderAdvancedRepoSearch(ctx *context.Context, opts *RepoSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		repos []*models.Repository
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0  && ctx.Query("professor") == "" && ctx.Query("subject") == "" && ctx.Query("group") == "" && ctx.Query("semester") == "" {
		repos, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter(opts.Private)
	} else {
		repos, count, err = models.AdvancedSearchRepositoryByName(&models.AdvancedSearchRepoOptions{
			Keyword:  keyword,
			Prof:     ctx.Query("professor"),
			Subj:     ctx.Query("subject"),
			Group:    ctx.Query("group"),
			Sem:      ctx.Query("semester"),
			OrderBy:  opts.OrderBy,
			Private:  opts.Private,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchRepositoryByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)

	for _, repo := range repos {
		if err = repo.GetOwner(); err != nil {
			log.Trace("%s", repo)
			ctx.Handle(500, "GetOwner", fmt.Errorf("%d: %v", repo.ID, err))
			return
		}
	}
	ctx.Data["Repos"] = repos

	ctx.HTML(200, opts.TplName)
}

func AdvancedSearch(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("advanced")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreAdvanced"] = true

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

	professors, err := models.GetProfessors()
	if err != nil {
		ctx.Handle(500, "GetProfessors", err)
		return
	}
	ctx.Data["Professors"] = professors

	RenderAdvancedRepoSearch(ctx, &RepoSearchOptions{
		Counter:  models.CountRepositories,
		Ranger:   models.GetRecentUpdatedRepositories,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "repository.updated_unix DESC",
		TplName:  EXPLORE_ADV,
	})
}

type UserSearchOptions struct {
	Type     models.UserType
	Counter  func() int64
	Ranger   func(int, int) ([]*models.User, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderUserSearch(ctx *context.Context, opts *UserSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		users []*models.User
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		users, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		users, count, err = models.SearchUserByName(&models.SearchUserOptions{
			Keyword:  keyword,
			Type:     opts.Type,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchUserByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Users"] = users

	ctx.HTML(200, opts.TplName)
}

func ExploreUsers(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreUsers"] = true

	RenderUserSearch(ctx, &UserSearchOptions{
		Type:     models.USER_TYPE_INDIVIDUAL,
		Counter:  models.CountUsers,
		Ranger:   models.Users,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_USERS,
	})
}

func ExploreProfessors(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreProfessors"] = true

	RenderUserSearch(ctx, &UserSearchOptions{
		Type:     models.USER_TYPE_PROFESSOR,
		Counter:  models.CountProfessors,
		Ranger:   models.Professors,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_PROFESSORS,
	})
}

func ExploreOrganizations(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreOrganizations"] = true

	RenderUserSearch(ctx, &UserSearchOptions{
		Type:     models.USER_TYPE_ORGANIZATION,
		Counter:  models.CountOrganizations,
		Ranger:   models.Organizations,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_ORGANIZATIONS,
	})
}

type SubjectsSearchOptions struct {
	Counter  func() int64
	Ranger   func(int, int) ([]*models.Subject, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderSubjectsSearch(ctx *context.Context, opts *SubjectsSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		subjects []*models.Subject
		count    int64
		err      error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		subjects, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		subjects, count, err = models.SearchSubjectByName(&models.SearchSubjectOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchSubjectByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Subjects"] = subjects

	ctx.HTML(200, opts.TplName)
}

func ExploreSubjects(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreSubjects"] = true

	RenderSubjectsSearch(ctx, &SubjectsSearchOptions{
		Counter:  models.CountSubjects,
		Ranger:   models.Subjects,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "name ASC",
		TplName:  EXPLORE_SUBJECTS,
	})
}

type SemestersSearchOptions struct {
	Counter  func() int64
	Ranger   func(int, int) ([]*models.Semester, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderSemestersSearch(ctx *context.Context, opts *SemestersSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		semesters []*models.Semester
		count     int64
		err       error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		semesters, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		semesters, count, err = models.SearchSemesterByName(&models.SearchSemesterOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchSemesterByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Semesters"] = semesters

	ctx.HTML(200, opts.TplName)
}

func ExploreSemesters(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreSemesters"] = true

	RenderSemestersSearch(ctx, &SemestersSearchOptions{
		Counter:  models.CountSemesters,
		Ranger:   models.Semesters,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "name ASC",
		TplName:  EXPLORE_SEMESTERS,
	})
}

type GrupsSearchOptions struct {
	Counter  func() int64
	Ranger   func(int, int) ([]*models.Group, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderGroupsSearch(ctx *context.Context, opts *GrupsSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		groups []*models.Group
		count  int64
		err    error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		groups, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		groups, count, err = models.SearchGroupByName(&models.SearchGroupOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchGroupByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Groups"] = groups

	ctx.HTML(200, opts.TplName)
}

func ExploreGroups(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreGroups"] = true

	RenderGroupsSearch(ctx, &GrupsSearchOptions{
		Counter:  models.CountGroups,
		Ranger:   models.Groups,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "name ASC",
		TplName:  EXPLORE_GROUPS,
	})
}

type TagsSearchOptions struct {
	Counter  func() int64
	Ranger   func(int, int) ([]*models.Tag, error)
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderTagsSearch(ctx *context.Context, opts *TagsSearchOptions) {
	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		tags  []*models.Tag
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		tags, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			ctx.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		tags, count, err = models.SearchTagByName(&models.SearchTagOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			ctx.Handle(500, "SearchTagByName", err)
			return
		}
	}
	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	ctx.Data["Tags"] = tags

	ctx.HTML(200, opts.TplName)
}

func ExploreTags(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreTags"] = true

	RenderTagsSearch(ctx, &TagsSearchOptions{
		Counter:  models.CountTags,
		Ranger:   models.Tags,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "etiqueta ASC",
		TplName:  EXPLORE_TAGS,
	})
}

func NotFound(ctx *context.Context) {
	ctx.Data["Title"] = "Page Not Found"
	ctx.Handle(404, "home.NotFound", nil)
}
