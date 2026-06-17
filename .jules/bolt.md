## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2026-06-17 - Avoid ErrUnsupportedDestination when doing fast existence checks
**Learning:** In GORM, when reading a scalar value into a primitive type (like `int`), use `.Scan(&dummy)` instead of `.Find(&dummy)`. Using `.Find()` with a primitive type will cause a runtime `ErrUnsupportedDestination` error, as it expects a struct or slice of structs.
**Action:** Use `.Select("1").Limit(1).Scan(&dummy)` instead of `Find()`.
