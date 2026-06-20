## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Scan(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables. Additionally, using `.Find()` to a primitive value throws an `ErrUnsupportedDestination` error, so `.Scan()` must be used.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Scan(&dummy)` pattern over `Count()`.
