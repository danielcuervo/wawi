# package httpc

`package httpc` provides a set of features for easy extension of the default `net/http` client.


## Decorators

The idea is pretty simple, start from a `http.DefaultClient` and decorate it with the functionality you need:

- Headers
- Instrumentations
- Fault tolerance (attempts with a backoff)
- Etc... 
 
## Example usage

See [example_test.go](example_test.go) file for example usages.

 
