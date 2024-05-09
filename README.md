# Bahamut

> Note: this is a work in progress.

Bahamut is a Go library that provides everything you need to set up a full blown
API server based on an [Elemental](https://go.aporeto.io/elemental) model
generated from a [Regolithe Specification](https://go.aporeto.io/regolithe).

The main concept of Bahamut is to only write core business logic, and letting it
handle all the boring bookkeeping. You can implement various Processors
interfaces, and register them when you start a Bahamut Server.

A Bahamut Server is not directly responsible for storing an retrieving data from
a database. To do so, you can use any backend library you like in your
processors, but we recommend using
[Manipulate](https://go.aporeto.io/manipulate), which provides a common
interface for manipulating an Elemental model and multiple implementations for
MongoDB (manipmongo), MemDB (manipmemory) or can be used to issue ReST calls using
maniphttp.

It is usually used by clients to interact with the API of a Bahamut service, but
also used for Bahamut Services to talk together.
