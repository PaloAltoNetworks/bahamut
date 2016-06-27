package processors

import (
	"github.com/satori/go.uuid"

	"github.com/aporeto-inc/bahamut"

	"github.com/aporeto-inc/bahamut/example/server/db"
	"github.com/aporeto-inc/bahamut/example/server/models"
)

// TaskProcessor processes lists
type TaskProcessor struct {
}

// NewTaskProcessor creates a new processor
func NewTaskProcessor() *TaskProcessor {

	return &TaskProcessor{}
}

// ProcessCreate does things.
func (p *TaskProcessor) ProcessCreate(ctx *bahamut.Context) error {

	task := ctx.InputData.(*models.Task)

	task.ID = uuid.NewV4().String()
	task.ParentID = ctx.Info.ParentIdentifier
	task.ParentType = ctx.Info.ParentIdentity.Name
	db.InsertTask(task)
	ctx.OutputData = task

	return nil
}

// ProcessRetrieveMany does things.
func (p *TaskProcessor) ProcessRetrieveMany(ctx *bahamut.Context) error {

	ret := db.TasksWithParentID(ctx.Info.ParentIdentifier)

	s, e := ctx.Page.IndexRange()

	if s > len(ret) {
		s = len(ret)
	}

	if e > len(ret) {
		e = len(ret)
	}

	ctx.OutputData = db.TasksInRange(s, e, ctx.Info.ParentIdentifier)
	ctx.Count.Current = e - s
	ctx.Count.Total = len(ret)

	return nil
}

// ProcessRetrieve does things.
func (p *TaskProcessor) ProcessRetrieve(ctx *bahamut.Context) error {

	ret, _, err := db.TaskWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	ctx.OutputData = ret

	return nil
}

// ProcessUpdate does things.
func (p *TaskProcessor) ProcessUpdate(ctx *bahamut.Context) error {

	orig, index, err := db.TaskWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	task := ctx.InputData.(*models.Task)
	task.ID = orig.ID
	task.ParentID = orig.ParentID
	db.UpdateTask(index, task)
	ctx.OutputData = task

	return nil
}

// ProcessDelete does things.
func (p *TaskProcessor) ProcessDelete(ctx *bahamut.Context) error {

	obj, index, err := db.TaskWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	db.DeleteTask(index)
	ctx.OutputData = obj
	return nil
}
