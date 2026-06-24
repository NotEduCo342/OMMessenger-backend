## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2024-06-25 - GORM Existence Checks using .Scan over .Find
**Learning:** In GORM, utilizing `.Select("1").Limit(1).Scan(&dummy)` provides significant performance over `.Count()` for existence checks because it short-circuits the query rather than scanning the entire table/index. Furthermore, when querying scalar values into a primitive type (like `int`), we must use `.Scan(&dummy)` rather than `.Find(&dummy)`, because `.Find` expects a struct or slice of structs and fails at runtime with `ErrUnsupportedDestination` for primitives.
**Action:** Always favor `.Scan` over `.Find` when querying scalar values directly into primitives, and employ the `.Select("1").Limit(1).Scan(&dummy)` pattern instead of `.Count()` for fast existence checks across the codebase.
