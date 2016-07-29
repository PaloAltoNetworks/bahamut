#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

monogen -f ./specs -L elemental

echo -n > ./model_test.go
cat ./codegen/elemental/1.0/list.go >> ./model_test.go
cat ./codegen/elemental/1.0/task.go | (read; read; read; read; cat) >> ./model_test.go
cat >> ./model_test.go << EOF

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
EOF

gofmt -w ./model_test.go
goimports -w ./model_test.go
mv ./model_test.go ../

rm -rf codegen
rm -f ./model_test.go-e
cd -
