#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

tg --name "ca" --is-ca --org "Aporeto" --common-name "localhost"
tg --name "client" --auth-server  --common-name "localhost" --signing-cert ./ca-cert.pem --signing-cert-key ./ca-key.pem
tg --name "server" --auth-server  --common-name "localhost" --signing-cert ./ca-cert.pem --signing-cert-key ./ca-key.pem

monogen -f ./specs -L elemental

echo -n > ./data_test.go
cat ./codegen/elemental/list.go >> ./data_test.go
cat ./codegen/elemental/task.go | (read; read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/root.go | (read; read; read; read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/user.go | (read; read;read; read; cat) >> ./data_test.go
cat ./codegen/elemental/relationships_registry.go | (read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/identities_registry.go | (read; read; read; cat) >> ./data_test.go
cat >> ./data_test.go << EOF

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

gofmt -w ./data_test.go
goimports -w ./data_test.go
mv ./data_test.go ../

rm -rf codegen
rm -f ./data_test.go-e
cd -
