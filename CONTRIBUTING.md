# Contributing to Handcarry-API

#### Table of Contents

[Coding Style](#coding-style)

[gqlgen](#gqlgen)

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

#### GraphQL error responses

Errors occurring at the top level of a query or mutation, such as an authorization
failure of an entire query, should present to the API client an http OK (200)
response with an `errors` object in the response body. Internally, this is handled
by the helper function `domain.ReportError`, which takes an `error` and a translation
key like `"UpdateUser.NotFound"`. The `error` is logged to `stderr` and to Rollbar,
and the translation key is localized by Buffalo's `Translate` function and returned
from the resolver to be processed by `gqlgen`. Translation text is stored in the
`application/locales` folder.

For example:

```go
extras := map[string]interface{}{
    "oldStatus": request.Status,
    "newStatus": input.Status,
}    
return nil, domain.ReportError(ctx, err, "UpdateRequestStatus.Unauthorized", extras)
``` 

Errors within a query, such as an authorization failure on a request field, should
not present an error to the API client. The requested data field should be returned
as `null`. The cause of the error should be logged to `stderr` and Rollbar using
`domain.ReportError` if context is available, or `domain.ErrLogger.Printf` if no
context is available.

#### REST API responses

Errors occurring in the processing of REST API requests should result in a 400-
or 500-level http response with a json body containing a `code` and a `key`:

```json
{
  "code": 400,
  "key": "ErrorMissingClientID"
}
``` 

This can be generated as follows:

```go
err := domain.AppError{
    Code: httpStatus,
    Key:  errorCode,
}

return c.Render(httpStatus, render.JSON(err))
```

In addition, the error should be logged using `Error` or `ErrLogger.Printf`. For 
auth-related errors, the helper `actions.authRequestError` is available.

#### Internal error logging

Errors that do not justify an error being passed to the API client may be logged
to `stderr` and Rollbar using `domain.Error` if context is available, or
`domain.ErrLogger.printf` if no context is available.
 
## gqlgen

gqlgen generates code to handle GraphQL queries. The primary input is the 
schema itself: [schema.graphql](application/gqlgen/schema.graphql). This file
syntax is standardized and [documented](#graphql-documentation). The other input
is the [gqlgen.yml](application/gqlgen/gqlgen.yml)` file.

### Add a query

It is desirable to keep the number of queries to a minimum, as there is little
room for structure and organization at the top level of the query hierarchy.
When appropriate, adding a query is as simple as adding a new field to the
`Query` type in [schema.graphql](application/gqlgen/schema.graphql) and
running the [generate](#generate) tool. This adds a new function to the
`queryResolver` interface. At this point, you need to define this new function
in the appropriate file in the `gqlgen`
package.  

### Add a mutation

This is similar to adding a query, except that you will add a field to the
`Mutation` type and add a function to the `mutationResolver` in `mutations.go`.
 
### Add a field to an existing query

Adding fields is the preferred method for extending the schema as it enhances
the query structure for future use. The process is similar: add one or more
fields to an existing type in [schema.graphql](application/gqlgen/schema.graphql)
and run the [generate](#generate) tool. This adds a function to the field-level
resolver (e.g. `userResolver`) interface. Again, just add the new function to a
file in the `gqlgen` package.

### Add an argument

Much of the power of GraphQL comes from query arguments. After adding an
argument and regenerating the code, you will see an additional argument in
the interface definition of the applicable resolver. Just add a new
argument to your function definition to match the corresponding interface.  

### Generate

To run the gqlgen code generator, execute `make gqlgen`. This runs`go generate
./...` inside a Docker container.

### GraphQL documentation

The graphql.org site has easy-to-read documentation on the schema format and 
other GraphQL information. The schema help is at
[https://graphql.org/learn/schema](https://graphql.org/learn/schema)

## Profiling with pprof

To use pprof for profiling WeCarry, it is available when `GO_ENV=development`. 
When running via `make` locally, `docker exec` into the buffalo container and
run commands like `go tool pprof  http://localhost:6060/debug/pprof/heap` to 
view the results. 

Good resources:

 - https://blog.gobuffalo.io/how-to-use-pprof-with-buffalo-983e5d71e418
 - https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/
