package models

import (
	"time"

	"github.com/apache/incubator-devlake/core/models/common"
	"gorm.io/datatypes"
)

type GithubProjectItem struct {
	ConnectionId  uint64 `gorm:"primaryKey"`
	ProjectItemId string `gorm:"primaryKey"`
	ProjectTitle  string
	ProjectNumber int
	Title         string
	Body          string
	Author        string
	IsArchived    bool
	Type          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ClosedAt      *time.Time
	Merged        *bool
	Assignees     *string           // Comma-separated list of assignee names
	Milestone     *string           // Milestone title
	Repository    *string           `gorm:"type:varchar(255)"` // Repository name
	DynamicFields datatypes.JSONMap `gorm:"type:json"`
	common.NoPKModel
}

func (GithubProjectItem) TableName() string {
	return "_tool_github_project_items"
}
