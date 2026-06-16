## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2024-06-16 - Fast Exists Query with Scan
**Learning:** While replacing `Count(&count)` with `.Select("1").Limit(1)` is great, using `.Find(&dummy)` when `dummy` is a primitive type like `int` causes an `ErrUnsupportedDestination` runtime error in GORM.
**Action:** When extracting a single scalar value like `1` into an `int`, you must use `.Scan(&dummy)` instead of `.Find(&dummy)`.
