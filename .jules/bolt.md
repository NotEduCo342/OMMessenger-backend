## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.
## 2024-05-24 - GORM Exist Check Optimization
**Learning:** In GORM, `.Count(&count)` on existence checks causes full table/index scans. However, when optimizing to `.Select("1").Limit(1)`, we must use `.Scan(&dummy)` instead of `.Find(&dummy)`. Using `.Find(&dummy)` with an `int` variable results in a runtime error `ErrUnsupportedDestination` because `.Find` expects a struct or slice of structs.
**Action:** Always use `.Scan(&dummy)` or `.Pluck()` when extracting primitive scalar values from optimized GORM queries.
