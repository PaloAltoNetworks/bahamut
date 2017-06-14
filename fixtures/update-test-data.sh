#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

openssl genrsa -aes256 -out ca-key.pem 4096 \
 && openssl req -new -x509 -days 10365 -key ca-key.pem -out ca.pem -sha256 -subj "/C=US/ST=CA/L=San Jose/O=Aporeto/CN=localhost"

openssl genrsa -out server-key.pem 4096 \
  && openssl req -subj "/CN=localhost" -sha256 -new -key server-key.pem -out server.req \
  && openssl x509 -req -days 10365 -sha256 -in server.req -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem \

openssl genrsa -out client-key.pem 4096 \
  && openssl req -subj "/O=aporeto.com/OU=SuperAdmin/CN=superadmin" -new -key client-key.pem -out client.req \
  && openssl x509 -req -days 10365 -sha256 -in client.req -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem

monogen -f ./specs -L elemental

echo -n > ./data_test.go
cat ./codegen/elemental/list.go >> ./data_test.go
cat ./codegen/elemental/task.go | (read; read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/root.go | (read; read; read; read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/user.go | (read; read;read; read; cat) >> ./data_test.go
cat ./codegen/elemental/identities_registry.go | (read; read; read; cat) >> ./data_test.go
cat ./codegen/elemental/relationships_registry.go | (read; read; read; cat) >> ./data_test.go
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
