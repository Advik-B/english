package types

// MaxCallDepth bounds how deeply English function calls may nest.
//
// Runaway recursion previously had no diagnostic: the AST evaluator grew the Go
// stack until the runtime aborted the process with an unrecoverable "stack
// overflow" fatal error, and the instruction VM grew its frame slice until the
// process hung or was killed. Both engines now raise a normal, catchable
// language-level error at this depth instead.
//
// It lives here rather than in either engine so that both share one limit and
// one message, and neither engine has to depend on the other.
const MaxCallDepth = 10000

// StackOverflowErrorType is the error type name programs can catch:
//
//	Try doing: ... on StackOverflowError: ...
const StackOverflowErrorType = "StackOverflowError"

// StackOverflowMessage is the shared message format used by both engines.
// Arguments: the depth limit, then the name of the function being called.
const StackOverflowMessage = "maximum call depth of %d exceeded while calling '%s'" +
	"\n  Hint: check for recursion that never reaches a base case"
