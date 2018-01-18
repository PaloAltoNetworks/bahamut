# Bahamut

Bahamut is a Go library that provides everything you need to set up a full blown API server based on an [Elemental](https://github.com/aporeto-inc/elemental) model generated from a [Monolithe Specification](https://github.com/aporeto-inc/monolithe).

The main concept of Bahamut is to only write core business logic, and letting it handle all the boring bookkeeping. You can implement various Processors interfaces, and register them when you start a Bahamut Server.

The included Monolithe plugin generates all the needed routes and handlers to reroute the client requests to the correct method of the correct processor. Those handlers will perform basic operations, like validating the request's data are valid and conform to the specifications. When your processor is finally called, you can be sure that all basic possible errors have been checked and that you can safely assume everything is ready to be stored, retrieved, or computed.

A Bahamut Server is not directly responsible for storing an retrieving data from a database. To do so, you can use any backend library you like in your processors, but we recommend using [Manipulate](https://github.com/aporeto-inc/manipulate), which provides a common interface for manipulating an Elemental model and multiple implementations for MongoDB, Cassandra or MemDB (with more to come). Later on, switching from Cassandra to MongoDB will be a no brainer.

> ALL THE REST MUST BE REWRITTEN
