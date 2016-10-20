# Bahamut

Bahamut is is Go library that provides everything you need to set up a full blown API server based on a [Elemental](https://github.com/aporeto-inc/elemental) model generated from some [Monolithe](https://github.com/aporeto-inc/monolithe) specifications.

The main concept of Bahamut is to only write core business logic, and letting it handle all the  boring bookkeeping. You can implement various Processors interfaces, and register them when you start a Bahamut Server.

The included Monolithe plugin generates all the needed routes and handlers to reroute the client requests to the correct method of the correct processor. Those handlers will perform basic operations, like validating the request's data are conform to the specifications, etc. When your processor is finally called, you can be sure that all basic errors has been previously checked and that you can safely assume everything is ready to be stored, retrieved, or whatever.

A Bahamut Server is not directly responsible for storing an retrieving data. To do so, you can use any backend library you like in your processors, but we recommend using Manipulate, which provides various a common interface for storing data and multiple implementations for MongoDB, Cassandra or MemDB (with more to come). Then switching from Cassandra to MongoDB is a no brainer.

An full example of a simple Todo List application can be found [here](https://github.com/aporeto-inc/bahamut-example)

## Prerequisites

Install Monolithe and the Bahamut plugin:

    $ pip install git+https://github.com/aporeto-inc/monolithe.git
    $ pip install 'git+ssh://git@github.com/aporeto-inc/bahamut.git#subdirectory=monolithe'

## Installation

To get the Bahamut Go Library, Run:

    $ go get github.com/aporeto-inc/bahamut

To install the Bahamut Routes and Handlers monolithe plugin, run:
