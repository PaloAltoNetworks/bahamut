#/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

echo
echo "\033[0;34m# Elemental\033[0;0m"
monogen -f specs -L elemental
rm -f ./server/models/*.go
cp codegen/elemental/1.0/*.go ./server/models
echo "\033[0;32m[success] generated Elemental installed in models \033[0;0m"

echo
echo "\033[0;34m# Bahamut\033[0;0m"
monogen -f specs -L bahamut
rm -f ./server/handlers/*.go
cp -a codegen/bahamut/1.0/handlers ./server
rm -f server/routes/*.go
cp -a codegen/bahamut/1.0/routes ./server
echo "\033[0;32m[success] generated Bahamut installed in server \033[0;0m"
echo

rm -rf codegen
