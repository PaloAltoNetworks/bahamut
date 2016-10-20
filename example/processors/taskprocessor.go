package processors

import (
	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/manipulate"

	"github.com/aporeto-inc/bahamut/example/db"
	"github.com/aporeto-inc/bahamut/example/models"
)

// TaskProcessor processes lists
type TaskProcessor struct {
	manipulator manipulate.TransactionalManipulator
}

// NewTaskProcessor creates a new processor
func NewTaskProcessor(manipulator manipulate.TransactionalManipulator) *TaskProcessor {

	return &TaskProcessor{
		manipulator: manipulator,
	}
}

// ProcessCreate does things.
func (p *TaskProcessor) ProcessCreate(ctx *bahamut.Context) error {

	t := ctx.InputData.(*models.Task)
	t.ParentID = ctx.Info.ParentIdentifier
	t.ParentType = ctx.Info.ParentIdentity.Name

	return db.Create(p.manipulator, ctx)
}

// ProcessRetrieveMany does things.
func (p *TaskProcessor) ProcessRetrieveMany(ctx *bahamut.Context) error {

	return db.RetrieveMany(p.manipulator, ctx, &models.TasksList{})
}

// ProcessRetrieve does things.
func (p *TaskProcessor) ProcessRetrieve(ctx *bahamut.Context) error {

	return db.Retrieve(p.manipulator, ctx, models.NewTask())
}

// ProcessUpdate does things.
func (p *TaskProcessor) ProcessUpdate(ctx *bahamut.Context) error {

	return db.Update(p.manipulator, ctx)
}

// ProcessDelete does things.
func (p *TaskProcessor) ProcessDelete(ctx *bahamut.Context) error {

	return db.Delete(p.manipulator, ctx, models.NewTask())
}
