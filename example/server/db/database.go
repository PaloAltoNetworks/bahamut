package db

import (
	"fmt"
	"sync"

	"github.com/aporeto-inc/bahamut/example/server/models"
)

var lists models.ListsList
var listsLock sync.Mutex

var tasks models.TasksList
var tasksLock sync.Mutex

var users models.UsersList
var usersLock sync.Mutex

func init() {

	lists = models.ListsList{}
	tasks = models.TasksList{}
	users = models.UsersList{}
}

/******************************
    Users
*******************************/

//ListsCount returns the total number of lists
func ListsCount() int {

	listsLock.Lock()
	defer listsLock.Unlock()

	return len(lists)

}

// ListWithID returns the list with the given ID
func ListWithID(identifier string) (*models.List, int, error) {

	listsLock.Lock()
	defer listsLock.Unlock()

	for i, o := range lists {
		if o.ID == identifier {
			return o, i, nil
		}
	}

	return nil, -1, fmt.Errorf("unable to find list with id %s", identifier)
}

// ListsInRange retuns the lists in the given range
func ListsInRange(s, e int) []*models.List {

	listsLock.Lock()
	defer listsLock.Unlock()

	return lists[s:e]
}

// InsertList Inserts the given list
func InsertList(list *models.List) {

	listsLock.Lock()
	defer listsLock.Unlock()

	lists = append(lists, list)
}

// UpdateList update the list at the given index
func UpdateList(index int, list *models.List) {

	listsLock.Lock()
	defer listsLock.Unlock()

	lists[index] = list
}

// DeleteList deletes the list at given index
func DeleteList(index int) {

	listsLock.Lock()
	defer listsLock.Unlock()

	lists = append(lists[:index], lists[index+1:]...)
}

/******************************
    Users
*******************************/

//UsersCount returns the total number of lists
func UsersCount() int {

	usersLock.Lock()
	defer usersLock.Unlock()

	return len(users)

}

// UserWithID returns the user with the given ID
func UserWithID(identifier string) (*models.User, int, error) {

	usersLock.Lock()
	defer usersLock.Unlock()

	for i, o := range users {
		if o.ID == identifier {
			return o, i, nil
		}
	}

	return nil, -1, fmt.Errorf("unable to find user with id %s", identifier)
}

// UsersInRange retuns the users in the given range
func UsersInRange(s, e int) []*models.User {

	usersLock.Lock()
	defer usersLock.Unlock()

	return users[s:e]
}

// InsertUser Inserts the given user
func InsertUser(user *models.User) {

	usersLock.Lock()
	defer usersLock.Unlock()

	users = append(users, user)
}

// UpdateUser update the user at the given index
func UpdateUser(index int, user *models.User) {

	usersLock.Lock()
	defer usersLock.Unlock()

	users[index] = user
}

// DeleteUser deletes the user at given index
func DeleteUser(index int) {

	usersLock.Lock()
	defer usersLock.Unlock()

	users = append(users[:index], users[index+1:]...)
}

/******************************
    Tasks
*******************************/

//TasksCount returns the total number of lists
func TasksCount() int {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	return len(tasks)

}

// TaskWithID returns the task with the given ID
func TaskWithID(identifier string) (*models.Task, int, error) {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	for i, o := range tasks {
		if o.ID == identifier {
			return o, i, nil
		}
	}

	return nil, -1, fmt.Errorf("unable to find task with id %s", identifier)
}

// TasksWithParentID returns the task with the given ID
func TasksWithParentID(identifier string) models.TasksList {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	ret := models.TasksList{}
	for _, o := range tasks {
		if o.ParentID == identifier {
			ret = append(ret, o)
		}
	}

	return ret
}

// TasksInRange retuns the lists in the given range
func TasksInRange(s, e int, parentID string) []*models.Task {

	return TasksWithParentID(parentID)[s:e]
}

// InsertTask Inserts the given list
func InsertTask(task *models.Task) {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	tasks = append(tasks, task)
}

// UpdateTask update the list at the given index
func UpdateTask(index int, task *models.Task) {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	tasks[index] = task
}

// DeleteTask deletes the task at given index
func DeleteTask(index int) {

	tasksLock.Lock()
	defer tasksLock.Unlock()

	tasks = append(tasks[:index], tasks[index+1:]...)
}
