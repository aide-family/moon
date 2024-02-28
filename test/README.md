# test/e2e

This is home to e2e tests used for presubmit, periodic, and postsubmit jobs.

Some of these jobs are merge-blocking, some are release-blocking.

## how to run

1. when you are in project home dir
```shell
go test ./test/...
```

2. when you are in test dir
```shell
go test ./e2e/...
```