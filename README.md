# Archivarius

An HTTP service that allows you to compress files into an archive and extract files from there. Tested on macOS.

### Compile and run

To compile run
`go build -o .bin/archivarius main.go`

No specific configuration is required

To run start the compiled binary
`.bin/archivarius`

### API

#### Synchronous

The service has two synchronous methods
`/api/v1/compress` and `/api/v1/extract`

###### Request

Both methods accept POST requests with the same JSON data structure
```
{
  "file": "path/to/archive.zip",
  "dir": "path/to/directory",
  "filter": "shell file name pattern",
  "limit": 1
}
```
 - `file` is the path to archive to work with.
 Create in case of "compress", and read in case of "extract"
 - `dir` is the path to directory to work with.
 To read files from in case of "compress", and to write files to in case of "extract"
 - `filter` is a filter to compress/extract files that match the pattern provided. The pattern is a shell file name pattern, e.g. `*.txt`. [details](https://pkg.go.dev/path/filepath#Match)
 `*` mathches any number of characters (non-separator),
 `?` matches any single character (non-separator).
 - `limit` is the max number of files to be compressed/extracted.
 For compression files are orderd by size (larger first) before processing.
 For extraction files are not sorted and are read in order they were written to the archive. If the archive was created by the service, files were written in order by size, therefore, larger files will be processed first.
 If "limit" is absent or is equal to "0" the default value of limit is assumed. For compression the default is 10, for extraction default is "unlimited"

###### Response

In case of success HTTP 200 code is returned with JSON response
```
{
  "status": "ok"
}
```

If error occured, non 2** HTTP code is returned with JSON response of a form
```
{
  "status": "nok",
  "message": "detailes about the error"
}
```

#### Async

##### Starting operations

The service has two endpoints for starting background operations
`/api/v1/compress/async` and `/api/v1/extract/async`

###### Request
Both accept POST requests with the same data as for sync methods

###### Response
Response is similar to sync methods, with addition of `session_id` fields, e.g.
```
{
  "session_id": "ecd5fd02-3a77-43c1-8e4e-58769742ad2a",
  "status": "ok"
}
```

##### Status of operation
###### Request
For probing status of the session the two endpoints support `GET` methods. Query parameter `session_id` must be provided, e.g.
```
GET http://localhost/api/v1/compress/async?session_id=ecd5fd02-3a77-43c1-8e4e-58769742ad2a
```

###### Response
Response of "get session" method has th following structure
```
{
  "status": "finished",
  "result": {
    "status_code": 200,
    "response": {
      "status": "ok"
    }
  }
}
```
Here:
 - `status` is the status of execution of session. Can be one of: `created`, `started`, `finished`
 - `result` is the structure containing the result of operaion for a finished session.
 - - `status_code` is what would have been a HTTP response code for sync method
 - - `response` is the result structure from sync method. It has `status` and optional `message` fields

##### Remove session
It is possible to remove a session when it is no longer needed. `DELETE` HTTP method is used for this. Query parameter `session_id` must be provided, e.g.
```
DELETE http://localhost/api/v1/compress/async?session_id=ecd5fd02-3a77-43c1-8e4e-58769742ad2a
```
If delete is never called, session will not be removed.

### Tests

To run tests provided execute
`go test ./test`
