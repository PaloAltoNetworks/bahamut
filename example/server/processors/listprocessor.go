package processors

import (
	"github.com/satori/go.uuid"

	"github.com/aporeto-inc/bahamut"

	"github.com/aporeto-inc/bahamut/example/server/db"
	"github.com/aporeto-inc/bahamut/example/server/models"
)

// ListProcessor processes lists
type ListProcessor struct {
}

// NewListProcessor creates a new processor
func NewListProcessor() *ListProcessor {

	return &ListProcessor{}
}

// ProcessCreate does things.
func (p *ListProcessor) ProcessCreate(ctx *bahamut.Context) error {

	list := ctx.InputData.(*models.List)

	list.ID = uuid.NewV4().String()
	db.InsertList(list)
	ctx.OutputData = list

	// test := models.NewUser()
	// test.FirstName = "john"
	// test.LastName = "bob"
	// ctx.EnqueueEvents(elemental.NewEvent(elemental.EventCreate, test))

	return nil
}

// ProcessRetrieveMany does things.
func (p *ListProcessor) ProcessRetrieveMany(ctx *bahamut.Context) error {

	s, e := ctx.Page.IndexRange()

	c := db.ListsCount()

	if s > c {
		s = c
	}

	if e > c {
		e = c
	}

	ctx.OutputData = db.ListsInRange(s, e)
	ctx.Count.Current = e - s
	ctx.Count.Total = c

	return nil
}

// ProcessRetrieve does things.
func (p *ListProcessor) ProcessRetrieve(ctx *bahamut.Context) error {

	ret, _, err := db.ListWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	ctx.OutputData = ret

	return nil
}

// ProcessUpdate does things.
func (p *ListProcessor) ProcessUpdate(ctx *bahamut.Context) error {

	orig, index, err := db.ListWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	list := ctx.InputData.(*models.List)
	list.ID = orig.ID
	list.ParentID = orig.ParentID

	db.UpdateList(index, list)
	ctx.OutputData = list

	return nil
}

// ProcessDelete does things.
func (p *ListProcessor) ProcessDelete(ctx *bahamut.Context) error {

	obj, index, err := db.ListWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	db.DeleteList(index)
	ctx.OutputData = obj
	return nil
}
