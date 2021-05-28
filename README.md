# Bahamut

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f8d3dbbc552b4c8abf8985425d25c338)](https://www.codacy.com/gh/PaloAltoNetworks/bahamut/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=PaloAltoNetworks/bahamut&amp;utm_campaign=Badge_Grade) [![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f8d3dbbc552b4c8abf8985425d25c338)](https://www.codacy.com/gh/PaloAltoNetworks/bahamut/dashboard?utm_source=github.com&utm_medium=referral&utm_content=PaloAltoNetworks/bahamut&utm_campaign=Badge_Coverage)

> README IS A WORK IN PROGRESS AS WE ARE WRITTING MORE DOCUMENTATION ABOUT THIS PACKAGE.

Bahamut is a Go library that provides everything you need to set up a full blown API server based on an [Elemental](https://go.aporeto.io/elemental) model generated from a [Regolithe Specification](https://go.aporeto.io/regolithe).

The main concept of Bahamut is to only write core business logic, and letting it handle all the boring bookkeeping. You can implement various Processors interfaces, and register them when you start a Bahamut Server.

A Bahamut Server is not directly responsible for storing an retrieving data from a database. To do so, you can use any backend library you like in your processors, but we recommend using [Manipulate](https://go.aporeto.io/manipulate), which provides a common interface for manipulating an Elemental model and multiple implementations for MongoDB, Cassandra or MemDB (with more to come). Later on, switching from Cassandra to MongoDB will be a no brainer.
