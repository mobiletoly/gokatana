# About

GoKatana is a set of libraries to simplify building web- (or micro-) services.

# Functionality overview

# katapp

Application support. It provides a common interface to handle web service being started
as a standalone application or as a part of a bigger system. It deals with some common tasks
such as application common error handling, logging, configuration, etc

# katcache

Cache support. It provides a common interface to work with cache storages. It creates a foundation
for caching data and also provides basic caching such as in-memory cache.

## kathttp

HTTP server support. It provides a common interface to start HTTP server, to handle requests,
supports graceful shutdown etc. It deals with some common tasks such as running http server,
compression/decompression of traffic, mapping from GoKatana application errors to HTTP status
codes etc. For now it is based on Echo framework, more frameworks can be added in the future.

Some additional features are:

#### Compression/decompression

There is a flag `server/compression` in the config file that allows to enable or disable compression
of response payload.

## kathttpc

HTTP client support. It provides a common interface to make HTTP requests, to handle responses,
supports retries, timeouts, etc. It deals with some common tasks such as making request with
JSON payload, mapping from HTTP status codes to GoKatana application errors etc.

## katpg

PostgreSQL support. It provides a common interface to work with PostgreSQL database (based
on performant PGX library). Some additional features are:

- Database migrations - a way to manage database schema changes
- Key/value cache - cache collections with expiration time support
- Leader - a leader election mechanism based on PostgreSQL advisory locks to select leader among
  multiple instances of the same service

## katredis

Redis support. It provides a common interface to work with Redis database as
a key-value cache.

## katsqlite

SQLite support. It provides a common interface to work with SQLite database.
Some additional features are:

- Database migrations - a way to manage database schema changes
- Key/value cache - cache collections with expiration time support


# Using sample templates 

The easiest approach to start using GoKatana is to use one of the sample templates:

https://github.com/mobiletoly/gokatana-samples
