package db

import (
	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/manipulate"
)

// Create is a helper for create processors.
func Create(m manipulate.TransactionalManipulator, ctx *bahamut.Context) error {

	o := ctx.InputData.(manipulate.Manipulable)

	if err := m.Create(nil, o); err != nil {
		return err
	}

	ctx.OutputData = o

	return nil
}

// RetrieveMany does things.
func RetrieveMany(m manipulate.TransactionalManipulator, ctx *bahamut.Context, dest interface{}) error {

	mctx := manipulate.NewContext()

	if !ctx.Info.ParentIdentity.IsEmpty() {
		mctx.Filter = manipulate.NewFilterComposer().
			WithKey("ParentID").Equals(ctx.Info.ParentIdentifier).
			Done()
	}

	if err := m.RetrieveMany(mctx, ctx.Info.ChildrenIdentity, dest); err != nil {
		return err
	}

	ctx.OutputData = dest

	return nil
}

// Retrieve does things.
func Retrieve(m manipulate.TransactionalManipulator, ctx *bahamut.Context, o manipulate.Manipulable) error {

	o.SetIdentifier(ctx.Info.ParentIdentifier)

	if err := m.Retrieve(nil, o); err != nil {
		return err
	}

	ctx.OutputData = o

	return nil
}

// Update does things.
func Update(m manipulate.TransactionalManipulator, ctx *bahamut.Context) error {

	o := ctx.InputData.(manipulate.Manipulable)
	o.SetIdentifier(ctx.Info.ParentIdentifier)

	if err := m.Update(nil, o); err != nil {
		return err
	}

	ctx.OutputData = o

	return nil
}

// Delete does things.
func Delete(m manipulate.TransactionalManipulator, ctx *bahamut.Context, o manipulate.Manipulable) error {

	o.SetIdentifier(ctx.Info.ParentIdentifier)

	if err := m.Delete(nil, o); err != nil {
		return err
	}

	ctx.OutputData = o

	return nil
}
