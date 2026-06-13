## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2026-06-13 - Fast Exists Query in GORM (re-applied)
**Learning:** The fast-fail existence check using `.Select("1").Limit(1).Find(&dummy)` instead of `Count()` in GORM wasn't applied everywhere in the codebase (e.g., `IsBlocked`, `IsBlocker`). It is a highly effective optimization for large tables to avoid full scans.
**Action:** Consistently use the `.Select("1").Limit(1).Find(&dummy)` pattern across all existence checks in the codebase instead of `Count()`, and review existing codebases to retroactively apply it.
