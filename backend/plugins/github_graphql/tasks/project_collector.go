/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"encoding/json"

	// "strings"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/merico-dev/graphql"
)

const RAW_PROJECTS_TABLE = "github_graphql_projects"

// QueryProject lists project items in a project
//
//	organization(login: "grafana") {
//		projectV2(number: 218) {
//			items(first: 50) {
//					totalCount
//					nodes {
//							id
//							createdAt
//					}
//			},
//			fields(first: 50) {
//				totalCount
//				nodes{
//			 	... on ProjectV2FieldCommon {
//					 name
//					 dataType
//			 	}
//				}
//			},
//		}
//	}
type QueryProject struct {
	RateLimit struct {
		Cost int
	}
	Organization struct {
		ProjectV2 struct {
			Fields struct {
				TotalCount int64
				Nodes      []Field
				PageInfo   api.GraphqlQueryPageInfo
			} `graphql:"fields(first: 100)"`
			Items struct {
				// Edges
				TotalCount graphql.Int
				Nodes      []ProjectItem
				PageInfo   api.GraphqlQueryPageInfo
			} `graphql:"items(first: 100, after: $cursor)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"organization(login: $login)"`
}

// ProjectItem is a GitHub project item
type ProjectItem struct {
	Content     ProjectV2ItemContent
	FieldValues FieldValues `graphql:"fieldValues(first: 100)"`
	ID          string
	IsArchived  bool
	Type        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Project     ProjectMeta
	Author      Author `graphql:"creator"`
}

type ProjectMeta struct {
	ID     string
	Number int
	Title  string
}

// ProjectV2ItemContent contains Content for a ProjectItem
type ProjectV2ItemContent struct {
	DraftIssue  DraftContent       `graphql:"... on DraftIssue"`
	Issue       IssueContent       `graphql:"... on Issue"`
	PullRequest PullRequestContent `graphql:"... on PullRequest"`
}

// DraftContent of the ProjectItem
type DraftContent struct {
	Title     *string
	Body      *string
	CreatedAt *time.Time
	Assignees *Assignees `graphql:"assignees(first: 10)"`
}

// IssuePrContent of the ProjectItem
type IssueContent struct {
	Title      *string
	Body       *string
	CreatedAt  *time.Time
	Assignees  *Assignees `graphql:"assignees(first: 10)"`
	Milestone  *Milestone
	ClosedAt   *time.Time
	Repository *Repository
}

// IssuePrContent of the ProjectItem
type PullRequestContent struct {
	Title      *string
	Body       *string
	CreatedAt  *time.Time
	Assignees  *Assignees `graphql:"assignees(first: 10)"`
	Milestone  *Milestone
	ClosedAt   *time.Time
	Merged     *bool
	Repository *Repository
}

// Milestone is a GitHub Milestone
type Milestone struct {
	Closed  bool
	Creator struct {
		User User `graphql:"... on User"`
	}
	DueOn     time.Time
	ClosedAt  time.Time
	CreatedAt time.Time
	State     string
	Title     string
}

// Assignees to the ProjectItem
type Assignees struct {
	PageInfo   api.GraphqlQueryPageInfo
	TotalCount int64
	Nodes      []User
}

// ProjectItemsWithFields ...
type ProjectItemsWithFields struct {
	Items  []ProjectItem
	Fields []Field
}

// FieldValues are the values of each Field of a ProjectItem
type FieldValues struct {
	PageInfo   api.GraphqlQueryPageInfo
	TotalCount int64
	Nodes      []FieldValue
}

// Field is a field on a ProjectItem
type Field struct {
	Common ProjectV2FieldCommon `graphql:"... on ProjectV2FieldCommon"`
}

// FieldValue is a value for a Field
type FieldValue struct {
	DateValue      ProjectV2ItemFieldDateValue         `graphql:"... on ProjectV2ItemFieldDateValue"`
	TextValue      ProjectV2ItemFieldTextValue         `graphql:"... on ProjectV2ItemFieldTextValue"`
	SelectValue    ProjectV2ItemFieldSingleSelectValue `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
	IterationValue ProjectV2ItemFieldIterationValue    `graphql:"... on ProjectV2ItemFieldIterationValue"`
	LabelsValue    ProjectV2ItemFieldLabelValue        `graphql:"... on ProjectV2ItemFieldLabelValue"`
	NumberValue    ProjectV2ItemFieldNumberValue       `graphql:"... on ProjectV2ItemFieldNumberValue"`
	ReviewerValue  ProjectV2ItemFieldReviewerValue     `graphql:"... on ProjectV2ItemFieldReviewerValue"`
	RepoValue      ProjectV2ItemFieldRepositoryValue   `graphql:"... on ProjectV2ItemFieldRepositoryValue"`
}

// ProjectV2ItemFieldRepositoryValue ...
type ProjectV2ItemFieldRepositoryValue struct {
	Repository Repository
	Field      CommonField
}

// Repository is a code repository
type Repository struct {
	Name string
}

// ProjectV2ItemFieldReviewerValue ...
type ProjectV2ItemFieldReviewerValue struct {
	Reviewers `graphql:"reviewers(first: 10)"`
	Field     CommonField
}

// Reviewers ...
type Reviewers struct {
	Nodes []Reviewer
}

// Reviewer ...
type Reviewer struct {
	User `graphql:"... on User"`
}

// Project item author
type Author struct {
	Login string
}

// ProjectV2ItemFieldNumberValue is a value for a Number field
type ProjectV2ItemFieldNumberValue struct {
	Number *float64
	Field  CommonField
}

// ProjectV2ItemFieldLabelValue is a value for a Labels field
type ProjectV2ItemFieldLabelValue struct {
	ProjectLabels `graphql:"labels(first: 10)"`
	Field         CommonField
}

// ProjectLabels ...
type ProjectLabels struct {
	Nodes []ProjectLabel
}

// ProjectLabel ...
type ProjectLabel struct {
	Name string
}

// ProjectV2ItemFieldIterationValue is a value for an Iteration field
type ProjectV2ItemFieldIterationValue struct {
	Title *string
	Field CommonField
}

// ProjectV2ItemFieldSingleSelectValue is a value for a SingleSelect field
type ProjectV2ItemFieldSingleSelectValue struct {
	Name  *string
	Field CommonField
}

// CommonField ...
type CommonField struct {
	Common ProjectV2FieldCommon `graphql:"... on ProjectV2FieldCommon"`
}

// ProjectV2ItemFieldTextValue is a value for a Text field
type ProjectV2ItemFieldTextValue struct {
	Text  *string
	Field CommonField
}

// ProjectV2ItemFieldDateValue is a value for a Date field
type ProjectV2ItemFieldDateValue struct {
	CreatedAt time.Time
	Date      *string
	UpdatedAt time.Time
	Field     CommonField
}

// ProjectV2FieldCommon is common to fields
type ProjectV2FieldCommon struct {
	Name     string
	DataType string
}

var CollectProjectsMeta = plugin.SubTaskMeta{
	Name:             "Collect Projects",
	EntryPoint:       CollectProjects,
	EnabledByDefault: false,
	Description:      "Collect GitHub Projects data from GraphQL API and save to raw table.",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CODE},
}

var _ plugin.SubTaskEntryPoint = CollectProjects

func CollectProjects(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*tasks.GithubTaskData)
	var err errors.Error

	apiCollector, err := api.NewStatefulApiCollector(api.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: map[string]interface{}{
			"ConnectionId":  data.Options.ConnectionId,
			"Owner":         data.Options.Owner,
			"ProjectNumber": data.Options.ProjectNumber,
		},
		Table: RAW_PROJECTS_TABLE,
	})
	if err != nil {
		return err
	}

	err = apiCollector.InitGraphQLCollector(api.GraphqlCollectorArgs{
		GraphqlClient: data.GraphqlClient,
		PageSize:      50,
		BuildQuery: func(reqData *api.GraphqlRequestData) (interface{}, map[string]interface{}, error) {
			query := &QueryProject{}
			if reqData == nil {
				return query, map[string]interface{}{}, nil
			}
			variables := map[string]interface{}{
				"number": graphql.Int(*data.Options.ProjectNumber),
				"login":  graphql.String(data.Options.Owner),
				"cursor": (*graphql.String)(reqData.Pager.SkipCursor),
			}
			return query, variables, nil
		},
		GetPageInfo: func(iQuery interface{}, args *api.GraphqlCollectorArgs) (*api.GraphqlQueryPageInfo, error) {
			query := iQuery.(*QueryProject)
			return &query.Organization.ProjectV2.Items.PageInfo, nil
		},
		ResponseParser: func(queryWrapper any) (messages []json.RawMessage, err errors.Error) {
			query := queryWrapper.(*QueryProject)
			items := query.Organization.ProjectV2.Items.Nodes
			for _, item := range items {
				messages = append(messages, errors.Must1(json.Marshal(item)))
			}
			return
		},
	})
	if err != nil {
		return err
	}

	return apiCollector.Execute()
}
