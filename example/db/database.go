package db

import (
	"github.com/aporeto-inc/bahamut/example/models"

	memdb "github.com/hashicorp/go-memdb"
)

// Schema is the memory db schema.
var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		models.ListIdentity.Category: &memdb.TableSchema{
			Name: models.ListIdentity.Category,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
			},
		},
		models.TaskIdentity.Category: &memdb.TableSchema{
			Name: models.TaskIdentity.Category,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
				"ParentID": &memdb.IndexSchema{
					Name:    "ParentID",
					Indexer: &memdb.StringFieldIndex{Field: "ParentID"},
				},
			},
		},
	},
}
