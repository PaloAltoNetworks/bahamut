# Bahamut

[![codecov](https://codecov.io/gh/aporeto-inc/bahamut/branch/master/graph/badge.svg?token=gMtfEkiWUa)](https://codecov.io/gh/aporeto-inc/bahamut)

> README IS A WORK IN PROGRESS AS WE ARE WRITTING MORE DOCUMENTATION ABOUT THIS PACKAGE.

Bahamut is a Go library that provides everything you need to set up a full blown API server based on an [Elemental](https://go.aporeto.io/elemental) model generated from a [Regolithe Specification](https://go.aporeto.io/regolithe).

The main concept of Bahamut is to only write core business logic, and letting it handle all the boring bookkeeping. You can implement various Processors interfaces, and register them when you start a Bahamut Server.

A Bahamut Server is not directly responsible for storing an retrieving data from a database. To do so, you can use any backend library you like in your processors, but we recommend using [Manipulate](https://go.aporeto.io/manipulate), which provides a common interface for manipulating an Elemental model and multiple implementations for MongoDB, Cassandra or MemDB (with more to come). Later on, switching from Cassandra to MongoDB will be a no brainer.
