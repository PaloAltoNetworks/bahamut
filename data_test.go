package bahamut

import "fmt"
import "github.com/aporeto-inc/elemental"

import "sync"

// ListIdentity represents the Identity of the object
var ListIdentity = elemental.Identity{
	Name:     "list",
	Category: "lists",
}

// ListsList represents a list of Lists
type ListsList []*List

// Copy returns a copy the TasksList.
func (o ListsList) Copy() elemental.ContentIdentifiable {

	return append(ListsList{}, o...)
}

// Append appends.
func (o ListsList) Append(objects ...elemental.Identifiable) elemental.ContentIdentifiable {

	out := append(ListsList{}, o...)
	for _, obj := range objects {
		out = append(out, obj.(*List))
	}

	return out
}

// ContentIdentity returns the identity of the objects in the list.
func (o ListsList) ContentIdentity() elemental.Identity {

	return ListIdentity
}

// List converts the object to an elemental.IdentifiablesList.
func (o ListsList) List() elemental.IdentifiablesList {

	out := elemental.IdentifiablesList{}
	for _, item := range o {
		out = append(out, item)
	}

	return out
}

// DefaultOrder returns the default ordering fields of the content.
func (o ListsList) DefaultOrder() []string {

	return []string{}
}

// Version returns the version
func (o ListsList) Version() int {

	return 1
}

// List represents the model of a list
type List struct {
	// The identifier
	ID string `json:"ID" bson:"_id"`

	// A creation only only attribute
	CreationOnly string `json:"creationOnly" bson:"creationonly"`

	// The description
	Description string `json:"description" bson:"description"`

	// The name
	Name string `json:"name" bson:"name"`

	// The identifier of the parent of the object
	ParentID string `json:"parentID" bson:"parentid"`

	// The type of the parent of the object
	ParentType string `json:"parentType" bson:"parenttype"`

	// A read only attribute
	ReadOnly string `json:"readOnly" bson:"readonly"`

	ModelVersion int `json:"-" bson:"_modelversion"`

	sync.Mutex
}

// NewList returns a new *List
func NewList() *List {

	return &List{
		ModelVersion: 1,
	}
}

// Identity returns the Identity of the object.
func (o *List) Identity() elemental.Identity {

	return ListIdentity
}

// Identifier returns the value of the object's unique identifier.
func (o *List) Identifier() string {

	return o.ID
}

// SetIdentifier sets the value of the object's unique identifier.
func (o *List) SetIdentifier(ID string) {

	o.ID = ID
}

// Version returns the hardcoded version of the model
func (o *List) Version() int {

	return 1
}

// DefaultOrder returns the list of default ordering fields.
func (o *List) DefaultOrder() []string {

	return []string{}
}

// Doc returns the documentation for the object
func (o *List) Doc() string {
	return `Represent a a list of task to do.`
}

func (o *List) String() string {

	return fmt.Sprintf("<%s:%s>", o.Identity().Name, o.Identifier())
}

// GetCreationOnly returns the creationOnly of the receiver
func (o *List) GetCreationOnly() string {
	return o.CreationOnly
}

// GetName returns the name of the receiver
func (o *List) GetName() string {
	return o.Name
}

// SetName set the given name of the receiver
func (o *List) SetName(name string) {
	o.Name = name
}

// GetReadOnly returns the readOnly of the receiver
func (o *List) GetReadOnly() string {
	return o.ReadOnly
}

// Validate valides the current information stored into the structure.
func (o *List) Validate() error {

	errors := elemental.Errors{}
	requiredErrors := elemental.Errors{}

	if err := elemental.ValidateRequiredString("creationOnly", o.CreationOnly); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("creationOnly", o.CreationOnly); err != nil {
		errors = append(errors, err)
	}

	if err := elemental.ValidateRequiredString("name", o.Name); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("name", o.Name); err != nil {
		errors = append(errors, err)
	}

	if err := elemental.ValidateRequiredString("readOnly", o.ReadOnly); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("readOnly", o.ReadOnly); err != nil {
		errors = append(errors, err)
	}

	if len(requiredErrors) > 0 {
		return requiredErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// SpecificationForAttribute returns the AttributeSpecification for the given attribute name key.
func (*List) SpecificationForAttribute(name string) elemental.AttributeSpecification {

	if v, ok := ListAttributesMap[name]; ok {
		return v
	}

	// We could not find it, so let's check on the lower case indexed spec map
	return ListLowerCaseAttributesMap[name]
}

// AttributeSpecifications returns the full attribute specifications map.
func (*List) AttributeSpecifications() map[string]elemental.AttributeSpecification {

	return ListAttributesMap
}

// ListAttributesMap represents the map of attribute for List.
var ListAttributesMap = map[string]elemental.AttributeSpecification{
	"ID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"CreationOnly": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		CreationOnly:   true,
		Description:    `A creation only only attribute`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "creationOnly",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"Description": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The description`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "description",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
	},
	"Name": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "name",
		Orderable:      true,
		Required:       true,
		Setter:         true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentType": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ReadOnly": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `A read only attribute`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "readOnly",
		Orderable:      true,
		ReadOnly:       true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

// ListLowerCaseAttributesMap represents the map of attribute for List.
var ListLowerCaseAttributesMap = map[string]elemental.AttributeSpecification{
	"id": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"creationonly": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		CreationOnly:   true,
		Description:    `A creation only only attribute`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "creationOnly",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"description": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The description`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "description",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
	},
	"name": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "name",
		Orderable:      true,
		Required:       true,
		Setter:         true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parentid": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parenttype": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"readonly": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `A read only attribute`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Getter:         true,
		Name:           "readOnly",
		Orderable:      true,
		ReadOnly:       true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

// TaskStatusValue represents the possible values for attribute "status".
type TaskStatusValue string

const (
	// TaskStatusDone represents the value DONE.
	TaskStatusDone TaskStatusValue = "DONE"

	// TaskStatusProgress represents the value PROGRESS.
	TaskStatusProgress TaskStatusValue = "PROGRESS"

	// TaskStatusTodo represents the value TODO.
	TaskStatusTodo TaskStatusValue = "TODO"
)

// TaskIdentity represents the Identity of the object
var TaskIdentity = elemental.Identity{
	Name:     "task",
	Category: "tasks",
}

// TasksList represents a list of Tasks
type TasksList []*Task

// Copy returns a copy the TasksList.
func (o TasksList) Copy() elemental.ContentIdentifiable {

	return append(TasksList{}, o...)
}

// Append appends.
func (o TasksList) Append(objects ...elemental.Identifiable) elemental.ContentIdentifiable {

	out := append(TasksList{}, o...)
	for _, obj := range objects {
		out = append(out, obj.(*Task))
	}

	return out
}

// ContentIdentity returns the identity of the objects in the list.
func (o TasksList) ContentIdentity() elemental.Identity {

	return TaskIdentity
}

// List converts the object to an elemental.IdentifiablesList.
func (o TasksList) List() elemental.IdentifiablesList {

	out := elemental.IdentifiablesList{}
	for _, item := range o {
		out = append(out, item)
	}

	return out
}

// DefaultOrder returns the default ordering fields of the content.
func (o TasksList) DefaultOrder() []string {

	return []string{}
}

// Version returns the version of the content.
func (o TasksList) Version() int {

	return 1
}

// Task represents the model of a task
type Task struct {
	// The identifier
	ID string `json:"ID" bson:"_id"`

	// The description
	Description string `json:"description" bson:"description"`

	// The name
	Name string `json:"name" bson:"name"`

	// The identifier of the parent of the object
	ParentID string `json:"parentID" bson:"parentid"`

	// The type of the parent of the object
	ParentType string `json:"parentType" bson:"parenttype"`

	// The status of the task
	Status TaskStatusValue `json:"status" bson:"status"`

	ModelVersion int `json:"-" bson:"_modelversion"`

	sync.Mutex
}

// NewTask returns a new *Task
func NewTask() *Task {

	return &Task{
		ModelVersion: 1,
		Status:       "TODO",
	}
}

// Identity returns the Identity of the object.
func (o *Task) Identity() elemental.Identity {

	return TaskIdentity
}

// Identifier returns the value of the object's unique identifier.
func (o *Task) Identifier() string {

	return o.ID
}

// SetIdentifier sets the value of the object's unique identifier.
func (o *Task) SetIdentifier(ID string) {

	o.ID = ID
}

// Version returns the hardcoded version of the model
func (o *Task) Version() int {

	return 1
}

// DefaultOrder returns the list of default ordering fields.
func (o *Task) DefaultOrder() []string {

	return []string{}
}

// Doc returns the documentation for the object
func (o *Task) Doc() string {
	return `Represent a task to do in a listd`
}

func (o *Task) String() string {

	return fmt.Sprintf("<%s:%s>", o.Identity().Name, o.Identifier())
}

// Validate valides the current information stored into the structure.
func (o *Task) Validate() error {

	errors := elemental.Errors{}
	requiredErrors := elemental.Errors{}

	if err := elemental.ValidateRequiredString("name", o.Name); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("name", o.Name); err != nil {
		errors = append(errors, err)
	}

	if err := elemental.ValidateStringInList("status", string(o.Status), []string{"DONE", "PROGRESS", "TODO"}, false); err != nil {
		errors = append(errors, err)
	}

	if len(requiredErrors) > 0 {
		return requiredErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// SpecificationForAttribute returns the AttributeSpecification for the given attribute name key.
func (*Task) SpecificationForAttribute(name string) elemental.AttributeSpecification {

	if v, ok := TaskAttributesMap[name]; ok {
		return v
	}

	// We could not find it, so let's check on the lower case indexed spec map
	return TaskLowerCaseAttributesMap[name]
}

// AttributeSpecifications returns the full attribute specifications map.
func (*Task) AttributeSpecifications() map[string]elemental.AttributeSpecification {

	return TaskAttributesMap
}

// TaskAttributesMap represents the map of attribute for Task.
var TaskAttributesMap = map[string]elemental.AttributeSpecification{
	"ID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"Description": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The description`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "description",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
	},
	"Name": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "name",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"ParentID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentType": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"Status": elemental.AttributeSpecification{
		AllowedChoices: []string{"DONE", "PROGRESS", "TODO"},
		DefaultValue:   TaskStatusValue("TODO"),
		Description:    `The status of the task`,
		Exposed:        true,
		Filterable:     true,
		Name:           "status",
		Orderable:      true,
		Stored:         true,
		Type:           "enum",
	},
}

// TaskLowerCaseAttributesMap represents the map of attribute for Task.
var TaskLowerCaseAttributesMap = map[string]elemental.AttributeSpecification{
	"id": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"description": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The description`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "description",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
	},
	"name": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "name",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"parentid": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parenttype": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"status": elemental.AttributeSpecification{
		AllowedChoices: []string{"DONE", "PROGRESS", "TODO"},
		DefaultValue:   TaskStatusValue("TODO"),
		Description:    `The status of the task`,
		Exposed:        true,
		Filterable:     true,
		Name:           "status",
		Orderable:      true,
		Stored:         true,
		Type:           "enum",
	},
}

// RootIdentity represents the Identity of the object
var RootIdentity = elemental.Identity{
	Name:     "root",
	Category: "root",
}

// Root represents the model of a root
type Root struct {
	// The identifier
	ID string `json:"ID" bson:"_id"`

	// The identifier of the parent of the object
	ParentID string `json:"parentID" bson:"parentid"`

	// The type of the parent of the object
	ParentType string `json:"parentType" bson:"parenttype"`

	Token        string `json:"APIKey,omitempty"`
	Organization string `json:"enterprise,omitempty"`
	ModelVersion int    `json:"-" bson:"_modelversion"`

	sync.Mutex
}

// NewRoot returns a new *Root
func NewRoot() *Root {

	return &Root{
		ModelVersion: 1,
	}
}

// Identity returns the Identity of the object.
func (o *Root) Identity() elemental.Identity {

	return RootIdentity
}

// Identifier returns the value of the object's unique identifier.
func (o *Root) Identifier() string {

	return o.ID
}

// SetIdentifier sets the value of the object's unique identifier.
func (o *Root) SetIdentifier(ID string) {

	o.ID = ID
}

// Version returns the hardcoded version of the model
func (o *Root) Version() int {

	return 1
}

// DefaultOrder returns the list of default ordering fields.
func (o *Root) DefaultOrder() []string {

	return []string{}
}

// Doc returns the documentation for the object
func (o *Root) Doc() string {
	return `Root object of the APIh`
}

func (o *Root) String() string {

	return fmt.Sprintf("<%s:%s>", o.Identity().Name, o.Identifier())
}

// Validate valides the current information stored into the structure.
func (o *Root) Validate() error {

	errors := elemental.Errors{}
	requiredErrors := elemental.Errors{}

	if len(requiredErrors) > 0 {
		return requiredErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// APIKey returns a the API Key
func (o *Root) APIKey() string {

	return o.Token
}

// SetAPIKey sets a the API Key
func (o *Root) SetAPIKey(key string) {

	o.Token = key
}

// SpecificationForAttribute returns the AttributeSpecification for the given attribute name key.
func (*Root) SpecificationForAttribute(name string) elemental.AttributeSpecification {

	if v, ok := RootAttributesMap[name]; ok {
		return v
	}

	// We could not find it, so let's check on the lower case indexed spec map
	return RootLowerCaseAttributesMap[name]
}

// AttributeSpecifications returns the full attribute specifications map.
func (*Root) AttributeSpecifications() map[string]elemental.AttributeSpecification {

	return RootAttributesMap
}

// RootAttributesMap represents the map of attribute for Root.
var RootAttributesMap = map[string]elemental.AttributeSpecification{
	"ID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentType": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

// RootLowerCaseAttributesMap represents the map of attribute for Root.
var RootLowerCaseAttributesMap = map[string]elemental.AttributeSpecification{
	"id": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parentid": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parenttype": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

// UserIdentity represents the Identity of the object
var UserIdentity = elemental.Identity{
	Name:     "user",
	Category: "users",
}

// UsersList represents a list of Users
type UsersList []*User

// Copy returns a copy the TasksList.
func (o UsersList) Copy() elemental.ContentIdentifiable {

	return append(UsersList{}, o...)
}

// Append appends.
func (o UsersList) Append(objects ...elemental.Identifiable) elemental.ContentIdentifiable {

	out := append(UsersList{}, o...)
	for _, obj := range objects {
		out = append(out, obj.(*User))
	}

	return out
}

// ContentIdentity returns the identity of the objects in the list.
func (o UsersList) ContentIdentity() elemental.Identity {

	return UserIdentity
}

// List converts the object to an elemental.IdentifiablesList.
func (o UsersList) List() elemental.IdentifiablesList {

	out := elemental.IdentifiablesList{}
	for _, item := range o {
		out = append(out, item)
	}

	return out
}

// DefaultOrder returns the default ordering fields of the content.
func (o UsersList) DefaultOrder() []string {

	return []string{}
}

// Version returns the default ordering fields of the content.
func (o UsersList) Version() int {

	return 1
}

// User represents the model of a user
type User struct {
	// The identifier
	ID string `json:"ID" bson:"_id"`

	// The first name
	FirstName string `json:"firstName" bson:"firstname"`

	// The last name
	LastName string `json:"lastName" bson:"lastname"`

	// The identifier of the parent of the object
	ParentID string `json:"parentID" bson:"parentid"`

	// The type of the parent of the object
	ParentType string `json:"parentType" bson:"parenttype"`

	// the login
	UserName string `json:"userName" bson:"username"`

	ModelVersion int `json:"-" bson:"_modelversion"`

	sync.Mutex
}

// NewUser returns a new *User
func NewUser() *User {

	return &User{
		ModelVersion: 1,
	}
}

// Identity returns the Identity of the object.
func (o *User) Identity() elemental.Identity {

	return UserIdentity
}

// Identifier returns the value of the object's unique identifier.
func (o *User) Identifier() string {

	return o.ID
}

// SetIdentifier sets the value of the object's unique identifier.
func (o *User) SetIdentifier(ID string) {

	o.ID = ID
}

// Version returns the hardcoded version of the model
func (o *User) Version() int {

	return 1
}

// DefaultOrder returns the list of default ordering fields.
func (o *User) DefaultOrder() []string {

	return []string{}
}

// Doc returns the documentation for the object
func (o *User) Doc() string {
	return `Represent a user.`
}

func (o *User) String() string {

	return fmt.Sprintf("<%s:%s>", o.Identity().Name, o.Identifier())
}

// Validate valides the current information stored into the structure.
func (o *User) Validate() error {

	errors := elemental.Errors{}
	requiredErrors := elemental.Errors{}

	if err := elemental.ValidateRequiredString("firstName", o.FirstName); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("firstName", o.FirstName); err != nil {
		errors = append(errors, err)
	}

	if err := elemental.ValidateRequiredString("lastName", o.LastName); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("lastName", o.LastName); err != nil {
		errors = append(errors, err)
	}

	if err := elemental.ValidateRequiredString("userName", o.UserName); err != nil {
		requiredErrors = append(requiredErrors, err)
	}

	if err := elemental.ValidateRequiredString("userName", o.UserName); err != nil {
		errors = append(errors, err)
	}

	if len(requiredErrors) > 0 {
		return requiredErrors
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// SpecificationForAttribute returns the AttributeSpecification for the given attribute name key.
func (*User) SpecificationForAttribute(name string) elemental.AttributeSpecification {

	if v, ok := UserAttributesMap[name]; ok {
		return v
	}

	// We could not find it, so let's check on the lower case indexed spec map
	return UserLowerCaseAttributesMap[name]
}

// AttributeSpecifications returns the full attribute specifications map.
func (*User) AttributeSpecifications() map[string]elemental.AttributeSpecification {

	return UserAttributesMap
}

// UserAttributesMap represents the map of attribute for User.
var UserAttributesMap = map[string]elemental.AttributeSpecification{
	"ID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"FirstName": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The first name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "firstName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"LastName": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The last name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "lastName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"ParentID": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"ParentType": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"UserName": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `the login`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "userName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

// UserLowerCaseAttributesMap represents the map of attribute for User.
var UserLowerCaseAttributesMap = map[string]elemental.AttributeSpecification{
	"id": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Identifier:     true,
		Name:           "ID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"firstname": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The first name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "firstName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"lastname": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `The last name`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "lastName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
	},
	"parentid": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The identifier of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		ForeignKey:     true,
		Format:         "free",
		Name:           "parentID",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"parenttype": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Autogenerated:  true,
		Description:    `The type of the parent of the object`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "parentType",
		Orderable:      true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
	"username": elemental.AttributeSpecification{
		AllowedChoices: []string{},
		Description:    `the login`,
		Exposed:        true,
		Filterable:     true,
		Format:         "free",
		Name:           "userName",
		Orderable:      true,
		Required:       true,
		Stored:         true,
		Type:           "string",
		Unique:         true,
	},
}

const nodocString = "[nodoc]" // nolint: varcheck

var relationshipsRegistry elemental.RelationshipsRegistry

// Relationships returns the model relationships.
func Relationships() elemental.RelationshipsRegistry {

	return relationshipsRegistry
}

func init() {
	relationshipsRegistry = elemental.RelationshipsRegistry{}

	relationshipsRegistry[ListIdentity] = &elemental.Relationship{
		AllowsCreate: map[string]bool{
			"root": true,
		},
		AllowsUpdate: map[string]bool{
			"root": true,
		},
		AllowsDelete: map[string]bool{
			"root": true,
		},
		AllowsRetrieve: map[string]bool{
			"root": true,
		},
		AllowsRetrieveMany: map[string]bool{
			"root": true,
		},
		AllowsInfo: map[string]bool{
			"root": true,
		},
	}
	relationshipsRegistry[TaskIdentity] = &elemental.Relationship{
		AllowsCreate: map[string]bool{
			"list": true,
		},
		AllowsUpdate: map[string]bool{
			"root": true,
		},
		AllowsDelete: map[string]bool{
			"root": true,
		},
		AllowsRetrieve: map[string]bool{
			"list": true,
			"root": true,
		},
		AllowsRetrieveMany: map[string]bool{
			"list": true,
			"root": true,
		},
		AllowsInfo: map[string]bool{
			"list": true,
			"root": true,
		},
	}
	relationshipsRegistry[RootIdentity] = &elemental.Relationship{}
	relationshipsRegistry[UserIdentity] = &elemental.Relationship{
		AllowsCreate: map[string]bool{
			"root": true,
		},
		AllowsUpdate: map[string]bool{
			"root": true,
			"list": true,
		},
		AllowsDelete: map[string]bool{
			"root": true,
		},
		AllowsRetrieve: map[string]bool{
			"list": true,
			"root": true,
		},
		AllowsRetrieveMany: map[string]bool{
			"list": true,
			"root": true,
		},
		AllowsInfo: map[string]bool{
			"list": true,
			"root": true,
		},
	}
}

func init() {

	elemental.RegisterIdentity(RootIdentity)
	elemental.RegisterIdentity(TaskIdentity)
	elemental.RegisterIdentity(ListIdentity)
	elemental.RegisterIdentity(UserIdentity)
}

// ModelVersion returns the current version of the model
func ModelVersion() int { return 1 }

// IdentifiableForIdentity returns a new instance of the Identifiable for the given identity name.
func IdentifiableForIdentity(identity string) elemental.Identifiable {

	switch identity {
	case RootIdentity.Name:
		return NewRoot()
	case TaskIdentity.Name:
		return NewTask()
	case ListIdentity.Name:
		return NewList()
	case UserIdentity.Name:
		return NewUser()
	default:
		return nil
	}
}

// IdentifiableForCategory returns a new instance of the Identifiable for the given category name.
func IdentifiableForCategory(category string) elemental.Identifiable {

	switch category {
	case RootIdentity.Category:
		return NewRoot()
	case TaskIdentity.Category:
		return NewTask()
	case ListIdentity.Category:
		return NewList()
	case UserIdentity.Category:
		return NewUser()
	default:
		return nil
	}
}

// ContentIdentifiableForIdentity returns a new instance of a ContentIdentifiable for the given identity name.
func ContentIdentifiableForIdentity(identity string) elemental.ContentIdentifiable {

	switch identity {
	case TaskIdentity.Name:
		return &TasksList{}
	case ListIdentity.Name:
		return &ListsList{}
	case UserIdentity.Name:
		return &UsersList{}
	default:
		return nil
	}
}

// ContentIdentifiableForCategory returns a new instance of a ContentIdentifiable for the given category name.
func ContentIdentifiableForCategory(category string) elemental.ContentIdentifiable {

	switch category {
	case TaskIdentity.Category:
		return &TasksList{}
	case ListIdentity.Category:
		return &ListsList{}
	case UserIdentity.Category:
		return &UsersList{}
	default:
		return nil
	}
}

// AllIdentities returns all existing identities.
func AllIdentities() []elemental.Identity {

	return []elemental.Identity{
		RootIdentity,
		TaskIdentity,
		ListIdentity,
		UserIdentity,
	}
}

var aliasesMap = map[string]elemental.Identity{}

// IdentityFromAlias returns the Identity associated to the given alias.
func IdentityFromAlias(alias string) elemental.Identity {

	return aliasesMap[alias]
}

// AliasesForIdentity returns all the aliases for the given identity.
func AliasesForIdentity(identity elemental.Identity) []string {

	switch identity {
	case RootIdentity:
		return []string{}
	case TaskIdentity:
		return []string{}
	case ListIdentity:
		return []string{}
	case UserIdentity:
		return []string{}
	}

	return nil
}

var UnmarshalableListIdentity = elemental.Identity{Name: "list", Category: "lists"}

type UnmarshalableList struct {
	List
}

func NewUnmarshalableList() *UnmarshalableList {
	return &UnmarshalableList{List: List{}}
}

func (o *UnmarshalableList) Identity() elemental.Identity { return UnmarshalableListIdentity }

func (o *UnmarshalableList) UnmarshalJSON([]byte) error {
	return fmt.Errorf("error unmarshalling")
}

func (o *UnmarshalableList) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("error marshalling")
}

func (o *UnmarshalableList) Validate() elemental.Errors { return nil }
