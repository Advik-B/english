# version

This module exists solely to store the version number of the English language interpreter.

## Version format

```
major.minor.commits-since-last-minor-increment
```

| Part | When to increment |
|------|-------------------|
| `major` | A change that **breaks** existing programs or behaviour |
| `minor` | A new **feature** is added |
| `commits-since-last-minor-increment` | Automatically derived — the number of git commits made since `minor` was last incremented |

## Usage

```go
import "github.com/Advik-B/english/version"

fmt.Println(version.Version) // e.g. "1.2.25"
```

> **Do not add any other logic to this package.** It is intentionally kept as a single constant so that every other package can import it without creating dependency cycles.
