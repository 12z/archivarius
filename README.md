# Archivarius

An HTTP service that allows you to compress files into an archive and extract files from there. Tested on macOS.

### Compile and run

To compile run
`go build -o .bin/archivarius main.go`

No specific configuration is required

To run start the compiled binary
`.bin/archivarius`

### API

The service has two methods
`/api/v1/compress` and `/api/v1/extract`

#### Request

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

#### Response

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

### Tests

To run tests provided execute
`go test ./test`
