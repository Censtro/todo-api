package tasks_postgres_repository

import (
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
)

type TaskModel struct {
	CreatedAt    time.Time
	Title        string
	ID           int
	Version      int
	CompletedAt  *time.Time
	AuthorUserID int
	Description  *string
	Completed    bool
}

func taskDomainFromModel(taskModel TaskModel) domain.Task {
	return domain.NewTask(
		taskModel.ID,
		taskModel.Version,
		taskModel.Title,
		taskModel.Description,
		taskModel.Completed,
		taskModel.CreatedAt,
		taskModel.CompletedAt,
		taskModel.AuthorUserID,
	)

}

func taskDomainsFromModels(taskModels []TaskModel) []domain.Task {
	domains := make([]domain.Task, len(taskModels))

	for i := range taskModels {
		domains[i] = domain.NewTask(
			taskModels[i].ID,
			taskModels[i].Version,
			taskModels[i].Title,
			taskModels[i].Description,
			taskModels[i].Completed,
			taskModels[i].CreatedAt,
			taskModels[i].CompletedAt,
			taskModels[i].AuthorUserID,
		)
	}

	return domains
}
