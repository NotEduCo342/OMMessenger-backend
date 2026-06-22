## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2025-02-12 - GORM primitive outputs
**Learning:** In GORM, when reading a scalar value into a primitive type (like int), use `.Scan(&dummy)` instead of `.Find(&dummy)`. Using `.Find()` with a primitive type will cause a runtime `ErrUnsupportedDestination` error, as it expects a struct or slice of structs.
**Action:** Use `.Scan()` when extracting simple scalar values like checking existence with `.Select("1")`.
