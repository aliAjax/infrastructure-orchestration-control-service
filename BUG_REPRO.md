# Bug Reproduction

## Bug

HTTP response wrappers overwrite the first committed status and do not reset or classify retry state consistently, causing client status, logging, and metrics to disagree.

## Trigger

Run:

```bash
go test ./api/middleware -run '^TestStableStatusWriterKeepsFirstErrorStatus$' -count=1
go test ./api/middleware -run '^TestLoggingStatusWriterKeepsFirstHeader$' -count=1
go test ./api/middleware -run '^TestStableStatusWriterWriteCommitsOK$' -count=1
go test ./api/middleware -run '^TestStableStatusWriterResetClearsWrittenState$' -count=1
go test ./api/middleware -run '^TestStableStatusWriterReportsErrorClass$' -count=1
go test ./api/middleware -run '^TestStableStatusWriterTracksRetryAttempt$' -count=1
```

## Error

```text
status_writer_test.go:15: status was overwritten: wrapper=500 recorder=404
status_writer_test.go:25: logging status changed after first header: 500
status_writer_test.go:37: body write did not commit 200: wrapper=500 recorder=200
status_writer_test.go:46: retry left the first-write latch set
status_writer_test.go:66: 5xx status was not classified as an error
status_writer_test.go:58: retry attempt was not advanced: 0
```
