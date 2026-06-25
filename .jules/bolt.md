## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2024-06-25 - GORM Fast-Fail Existence Checks and Runtime Errors
**Learning:** When optimizing GORM queries for existence checks using the fast-fail pattern (`.Select("1").Limit(1)`), it is crucial to use `.Scan(&dummy)` instead of `.Find(&dummy)` when reading into a primitive type (like `int`). Using `.Find()` expects a struct or a slice of structs, and using it with a scalar type results in a runtime `ErrUnsupportedDestination` error.
**Action:** Always use `.Scan(&dummy)` when applying the `.Select("1").Limit(1)` optimization pattern for checking existence with scalar variables.
