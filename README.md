# Database Query Executer Engine and ElasticSearch Bulk Indexing
This is a set of API'S made in golang used for making DB query and pushing those results to ElasticSearch through bulk indexing mechanism.

## Features

- Performs multiple indexing or delete operations in a single API call
- Managed Access Control for database operation
- Maintained a Standard logging mechanism for HTTP request and other logging scenarios
- Centralized Error handling and response mechanism
- It supports Mysql and Postgresql Databases
- Reusability of public sqltojson module for converting the sql data to json format

## Tech

Uses a number of open source projects to work properly:

- Golang 
- Elastic Search
- [sqltojson] module
- postgresql/mysql

## Installation

This requires [Golang] v1.0+ to run.

This will install the dependencies and run the go server at port 8084.

```sh
cd golang-db_query_engine-elasticsearch-indexer
go run main.go
```

## Usage
For Database query
```sh
curl --location --request GET 'localhost:8084/query/select * from table'
```

For bulk indexing data to ElasticSearch, this will create index if it does not exist
```sh
curl --location --request POST 'http://localhost:8084/elastic/save/?dbDriver=mysql&dbName=testdb&query=SELECT * FROM table&indexName=testindex'
```

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [sqltojson]: <https://github.com/rakeshkhetwal/sqltojson>
   [golang]: <https://go.dev/dl/>
