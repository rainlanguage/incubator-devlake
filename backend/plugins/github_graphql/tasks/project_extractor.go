package tasks

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/araddon/dateparse"
	"gorm.io/datatypes"
)

var _ plugin.SubTaskEntryPoint = ExtractProjects

var ExtractProjectsMeta = plugin.SubTaskMeta{
	Name:             "Extract Projects",
	EntryPoint:       ExtractProjects,
	EnabledByDefault: false,
	Description:      "Extract raw GitHub Project items into tool layer table github_project_items (with dynamic fields as JSON)",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CODE},
}

func ExtractProjects(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*tasks.GithubTaskData)

	extractor, err := api.NewApiExtractor(api.ApiExtractorArgs{
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: map[string]interface{}{
				"ConnectionId":  data.Options.ConnectionId,
				"Owner":         data.Options.Owner,
				"ProjectNumber": data.Options.ProjectNumber,
			},
			Table: RAW_PROJECTS_TABLE,
		},
		Extract: func(row *api.RawData) ([]interface{}, errors.Error) {
			rawItem := &ProjectItem{}
			err := errors.Convert(json.Unmarshal(row.Data, rawItem))
			if err != nil {
				return nil, err
			}

			// collect dynamic fields into a map
			dynFields := map[string]interface{}{}
			for _, fv := range rawItem.FieldValues.Nodes {
				name, val := nameValue(fv)
				dynFields[name] = val
			}

			project := &models.GithubProjectItem{
				ConnectionId:  data.Options.ConnectionId,
				ProjectItemId: rawItem.ID,
				ProjectTitle:  rawItem.Project.Title,
				ProjectNumber: rawItem.Project.Number,
				Title:         getProjectItemTitle(rawItem),
				Body:          getProjectItemBody(rawItem),
				Author:        rawItem.Author.Login,
				IsArchived:    rawItem.IsArchived,
				Type:          rawItem.Type,
				CreatedAt:     rawItem.CreatedAt,
				UpdatedAt:     rawItem.UpdatedAt,
				ClosedAt:      closedDate(rawItem.Content),
				Merged:        isMerged(rawItem.Content),
				Repository:    getProjectItemRepo(rawItem),
				Assignees:     getAssignees(rawItem.Content),
				Milestone:     milestone(rawItem.Content),
				DynamicFields: datatypes.JSONMap(dynFields),
			}

			return []interface{}{project}, nil
		},
	})
	if err != nil {
		return err
	}

	return extractor.Execute()
}

// Helper functions to extract title/body from ProjectItem content
func getProjectItemTitle(item *ProjectItem) string {
	if item.Content.DraftIssue.Title != nil {
		return *item.Content.DraftIssue.Title
	}
	if item.Content.Issue.Title != nil {
		return *item.Content.Issue.Title
	}
	if item.Content.PullRequest.Title != nil {
		return *item.Content.PullRequest.Title
	}
	return ""
}

// getProjectItemBody extracts the body from the ProjectItem content
func getProjectItemBody(item *ProjectItem) string {
	if item.Content.DraftIssue.Body != nil {
		return *item.Content.DraftIssue.Body
	}
	if item.Content.Issue.Body != nil {
		return *item.Content.Issue.Body
	}
	if item.Content.PullRequest.Body != nil {
		return *item.Content.PullRequest.Body
	}
	return ""
}

func getProjectItemRepo(item *ProjectItem) *string {
	if item.Content.Issue.Repository != nil {
		return &item.Content.Issue.Repository.Name
	}
	if item.Content.PullRequest.Repository != nil {
		return &item.Content.PullRequest.Repository.Name
	}
	return nil
}

// convert fieldValue to time
func date(fv FieldValue) *time.Time {
	if fv.DateValue.Date == nil {
		return nil
	}
	t, err := dateparse.ParseAny(*fv.DateValue.Date)
	if err != nil {
		return nil
	}
	return &t
}

// get closed date from the content
func closedDate(content ProjectV2ItemContent) *time.Time {
	if content.Issue.ClosedAt != nil {
		return content.Issue.ClosedAt
	}
	if content.PullRequest.ClosedAt != nil {
		return content.PullRequest.ClosedAt
	}
	return nil
}

// convert list of assignees to comma delimited string
func getAssignees(content ProjectV2ItemContent) *string {
	// if content.Issue.CreatedAt != nil {
	// 	return assignees(content.Issue)
	// }
	if content.Issue.CreatedAt != nil {
		return assignees(content.Issue.Assignees)
	}
	if content.DraftIssue.CreatedAt != nil {
		return assignees(content.DraftIssue.Assignees)
	}
	if content.PullRequest.CreatedAt != nil {
		return assignees(content.PullRequest.Assignees)
	}
	return nil
}

func assignees(assignees *Assignees) *string {
	var _assignees []string
	for _, v := range assignees.Nodes {
		_assignees = append(_assignees, v.Name)
	}
	names := strings.Join(_assignees, ",")
	return &names
}

// get milestone as string from the model
func milestone(content ProjectV2ItemContent) *string {
	if content.Issue.CreatedAt != nil && content.Issue.Milestone != nil {
		return &content.Issue.Milestone.Title
	}
	if content.PullRequest.CreatedAt != nil && content.PullRequest.Milestone != nil {
		return &content.PullRequest.Milestone.Title
	}
	return nil
}

// convert list of labels to comma delimited string
func labels(fv FieldValue) *string {
	if fv.LabelsValue.Nodes != nil {
		var labels []string
		for _, l := range fv.LabelsValue.Nodes {
			labels = append(labels, l.Name)
		}
		val := strings.Join(labels, ",")
		return &val
	}
	return nil
}

// convert list of reviewers to comma delimited string
func reviewers(fv FieldValue) *string {
	if fv.ReviewerValue.Nodes != nil {
		var vals []string
		for _, r := range fv.ReviewerValue.Nodes {
			vals = append(vals, r.Name)
		}
		val := strings.Join(vals, ",")
		return &val
	}
	return nil
}

func isMerged(content ProjectV2ItemContent) *bool {
	if content.PullRequest.Merged != nil {
		return content.PullRequest.Merged
	}
	return nil
}

// get the field name and value from the response model
func nameValue(fv FieldValue) (string, any) {
	// DateValue, SelectValue, TextValue etc all have the field data type
	dataType := fv.DateValue.Field.Common.DataType
	switch dataType {
	case "DATE":
		return fv.DateValue.Field.Common.Name, date(fv)
	case "SINGLE_SELECT":
		return fv.SelectValue.Field.Common.Name, fv.SelectValue.Name
	case "TEXT", "TITLE":
		return fv.TextValue.Field.Common.Name, fv.TextValue.Text
	case "ITERATION":
		return fv.IterationValue.Field.Common.Name, fv.IterationValue.Title
	case "LABELS":
		return fv.LabelsValue.Field.Common.Name, labels(fv)
	case "NUMBER":
		return fv.NumberValue.Field.Common.Name, fv.NumberValue.Number
	case "REVIEWERS":
		return fv.ReviewerValue.Field.Common.Name, reviewers(fv)
	case "REPOSITORY":
		return fv.RepoValue.Field.Common.Name, &fv.RepoValue.Repository.Name
	}
	return fv.DateValue.Field.Common.Name, nil
}
