## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2024-06-29 - GORM Primitive Type Fast Exists
**Learning:** Using `.Find(&dummy)` in GORM with a primitive type like `int` causes a runtime `ErrUnsupportedDestination` error. The `.Scan(&dummy)` method should be used when reading scalar values.
**Action:** Use `.Select("1").Limit(1).Scan(&dummy)` instead of `.Count()` or `.Find()` for fast-fail existence checks.
