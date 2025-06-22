package migrationscripts

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"github.com/apache/incubator-devlake/plugins/github/models"
)

type addGithubProjectItems20240613 struct{}

func (*addGithubProjectItems20240613) Up(basicRes context.BasicRes) errors.Error {
	return migrationhelper.AutoMigrateTables(
		basicRes,
		&models.GithubProjectItem{},
	)
}

func (*addGithubProjectItems20240613) Version() uint64 {
	return 20240613000000
}

func (*addGithubProjectItems20240613) Name() string {
	return "add table _tool_github_project_items"
}
