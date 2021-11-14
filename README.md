# Deliveroo Technical Test
I time-boxed myself 3 hours to create the solution (time spent writing readme and committing to this 
throwaway Github account not included).

## Install Go
The test description specifies this project will be run on a clean build of OS X/Linux. Assuming
`go` is not installed on said system, here is how to install it.

### OS X
```
$ brew install go
```

### Linux, Windows (or non homebrew OS X)
https://golang.org/doc/install :)

### Validating install
```
$ go version
```

A fresh install should get you `v1.17.3` or later.

## How to run
Simplest way to run:
```
$ go run main.go "*/15 0 1,15 * 1-5 /usr/bin/find"
```

### Compile and run
```
$ go build -o cronparser
$ ./cronparser "*/15 0 1,15 * 1-5 /usr/bin/find"
```

### Execute tests
```
$ go test -v ./...
```

## Potential improvements
- Better document the code, again due to time restraints I opted for functionality first
- Write tests to cover failures on specific parsers and validate correct error messages
- Extend to work with more than 5 time fields
- Implement extension functionality i.e, `@yearly`, `@monthly`, `@daily` etc.

## Source of CRON documentation
- https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format
- https://man7.org/linux/man-pages/man5/crontab.5.html

I chose to follow the IBM documentation as that closer resembled the Cron specification outlined in the
test case whereas `man cron` also documented extensions such as `@yearly`.