## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2023-10-25 - GORM Fast-Fail Existence Checks & Primitive Type Bug
**Learning:** Using `.Count()` in GORM causes a full table or index scan, which is very inefficient for simple existence checks. A known optimization pattern is using `.Select("1").Limit(1).Scan(&dummy)`, which returns fast. However, it is CRITICAL to use `.Scan()` instead of `.Find()` when scanning into a primitive type (like `int`), because `.Find()` will cause a runtime `ErrUnsupportedDestination` error.
**Action:** When implementing existence checks, always use the fast-fail pattern `.Select("1").Limit(1).Scan(&dummy)`. Never use `.Find()` for primitive scalar values in GORM.
