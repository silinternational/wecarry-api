# Contributing to Handcarry-API

#### Table of Contents

[Coding Style](#coding-style)

## Coding Style

### Go formatting

Because Go has one code formatting standard, this project uses that
standard. To stay consistent, enable `goimports` in your editor or IDE to
format your code before it's committed. For example, in Goland, go to Settings -
Tools - File Watchers, add and enable `goimports`.

### Function naming

Within the `model` package, we have decided on function names starting with
certain standardized verbs: Get, Find, Create, Delete. When possible, functions
should have a model struct attached as a pointer: `func (r *Request)
FindByUUID(uuid string) error`.

### Unit test naming

Unit test functions that test methods should be named like
`TestObject_FunctionName` where `Object` is the name of the struct and
`FunctionName` is the name of the function under test.

### Test suites

Use Buffalo ([strechr/testify](https://github.com/stretchr/testify)) test
suites. If not all tests in a package that uses Buffalo suites use the correct
syntax, then running `buffalo test -m TestObject_FunctionName` will run the
expected test and any standard Go test functions/suites. For example, since the
`models` package has a `models_test` suite, all tests in this package should be
of the form:
```go
func (ms *ModelSuite) TestObject_FunctionName() {
}
```
rather than  
```go
func Test_FunctionName(t *testing.T) {
}
```

### Running tests manually

To run all tests, run `make test`.

To run a single test:
1. run `make testenv` - this starts the test container and drops you into a bash prompt, from which you can run test commands.
2. `go test -v ./actions -testify.m "Test_Name"` - this runs more quickly than `buffalo test actions -m "Test_Name"` and allows you to use go test flags like `-v`.

### Database Queries

For simple queries and simple joins, Pop provides a good API based on
model struct annotations. These should be used where possible. Do not assume,
however, that objects passed from other functions are pre-populated with
data from related objects. If related data is required, call the `DB.Load`
function.

Complex queries and joins can be accomplished using the model fields and 
iterating over the attached lists. This ends up being more complex and 
difficult to read. We have determined it is better to use raw SQL in these
situations. For example:

```go
    var t Threads
    query := DB.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id")
    query = query.Where("tp.user_id = ?", u.ID)
    if err := query.All(&t); err != nil {
        return nil, err
    }
```
     
### Error handling and presentation

#### REST API responses

Errors occurring in the processing of REST API requests should result in a 400-
or 500-level http response with a json body like:

```json
{
  "code": 400,
  "key": "ErrorKeyExample",
  "message": "This is an example error message"
}
``` 

The type `api.AppError` will render as required above by passing it to 
`actions.reportError`. An `AppError` should be created by calling
`api.NewAppError` as deep into the call stack as needed to provide a detailed
key and specific category. If `actions.reportError` receives a generic `error`,
it will render with key `UnknownError` and HTTP status 500 and the error string
in the `DebugMsg`.

| Category          | HTTP Status |
|-------------------|-------------|
| CategoryInternal  | 500         |
| CategoryDatabase  | 500         |
| CategoryForbidden | 404         |
| CategoryNotFound  | 404         |
| CategoryUser      | 400         |

#### Internal error logging

Errors that do not justify an error being passed to the API client may be logged
to `stderr` and Rollbar using `domain.Error` if context is available, or
`domain.ErrLogger.printf` if no context is available.

`domain.Warn` can be used to log at level "warning" and also send to Rollbar

`domain.Info` or `domain.Logger.printf` will log but not send to Rollbar.
 
## Profiling with pprof

To use pprof for profiling WeCarry, it is available when `GO_ENV=development`. 
When running via `make` locally, `docker exec` into the buffalo container and
run commands like `go tool pprof  http://localhost:6060/debug/pprof/heap` to 
view the results. 

Good resources:

 - https://blog.gobuffalo.io/how-to-use-pprof-with-buffalo-983e5d71e418
 - https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/

## Debugging with Delve

Remote debugging with a compatible IDE is possible using the `delve` container. It does not have buffalo file watching capability, so any code changes will not be possible without a rebuild.

Set up in GoLand is as simple as adding a Run/Debug Configuration. Use type "Go Remote" and use default settings (host: localhost, port: 2345, on disconnect: ask).

To begin debugging, run `make debug`. This kills the `buffalo` container and starts the `debug` container. Once the app build is finished, click the debug button on the GoLand toolbar.
