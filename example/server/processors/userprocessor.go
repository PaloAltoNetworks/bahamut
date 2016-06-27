package processors

import (
	"fmt"

	"github.com/satori/go.uuid"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"

	"github.com/aporeto-inc/bahamut/example/server/db"
	"github.com/aporeto-inc/bahamut/example/server/models"
)

// UserProcessor processes users
type UserProcessor struct {
}

// NewUserProcessor creates a new processor
func NewUserProcessor() *UserProcessor {

	return &UserProcessor{}
}

// ProcessCreate does things.
func (p *UserProcessor) ProcessCreate(ctx *bahamut.Context) error {

	user := ctx.InputData.(*models.User)

	user.ID = uuid.NewV4().String()
	db.InsertUser(user)
	ctx.OutputData = user

	return nil
}

// ProcessRetrieveMany does things.
func (p *UserProcessor) ProcessRetrieveMany(ctx *bahamut.Context) error {

	s, e := ctx.Page.IndexRange()

	c := db.UsersCount()

	if s > c {
		s = c
	}

	if e > c {
		e = c
	}

	ctx.OutputData = db.UsersInRange(s, e)
	ctx.Count.Current = e - s
	ctx.Count.Total = c

	return nil
}

// ProcessRetrieve does things.
func (p *UserProcessor) ProcessRetrieve(ctx *bahamut.Context) error {

	ret, _, err := db.UserWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	ctx.OutputData = ret

	return nil
}

// ProcessUpdate does things.
func (p *UserProcessor) ProcessUpdate(ctx *bahamut.Context) error {

	orig, index, err := db.UserWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	user := ctx.InputData.(*models.User)
	user.ID = orig.ID
	user.ParentID = orig.ParentID

	db.UpdateUser(index, user)
	ctx.OutputData = user

	return nil
}

// ProcessDelete does things.
func (p *UserProcessor) ProcessDelete(ctx *bahamut.Context) error {

	obj, index, err := db.UserWithID(ctx.Info.ParentIdentifier)

	if err != nil {
		return err
	}

	db.DeleteUser(index)
	ctx.OutputData = obj
	return nil
}

// ProcessPatch does things.
func (p *UserProcessor) ProcessPatch(ctx *bahamut.Context) error {
	assignation := ctx.InputData.(*elemental.Assignation)
	fmt.Println(assignation)
	return nil
}
