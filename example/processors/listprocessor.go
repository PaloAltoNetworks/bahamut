package processors

import (
	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/manipulate"

	"github.com/aporeto-inc/bahamut/example/db"
	"github.com/aporeto-inc/bahamut/example/models"
)

// ListProcessor processes lists
type ListProcessor struct {
	manipulator manipulate.TransactionalManipulator
}

// NewListProcessor creates a new processor
func NewListProcessor(manipulator manipulate.TransactionalManipulator) *ListProcessor {

	return &ListProcessor{
		manipulator: manipulator,
	}
}

// ProcessCreate does things.
func (p *ListProcessor) ProcessCreate(ctx *bahamut.Context) error {

	return db.Create(p.manipulator, ctx)
}

// ProcessRetrieveMany does things.
func (p *ListProcessor) ProcessRetrieveMany(ctx *bahamut.Context) error {

	return db.RetrieveMany(p.manipulator, ctx, &models.ListsList{})
}

// ProcessRetrieve does things.
func (p *ListProcessor) ProcessRetrieve(ctx *bahamut.Context) error {

	return db.Retrieve(p.manipulator, ctx, models.NewList())
}

// ProcessUpdate does things.
func (p *ListProcessor) ProcessUpdate(ctx *bahamut.Context) error {

	return db.Update(p.manipulator, ctx)
}

// ProcessDelete does things.
func (p *ListProcessor) ProcessDelete(ctx *bahamut.Context) error {

	return db.Delete(p.manipulator, ctx, models.NewList())
}
