## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.
## 2025-02-14 - Optimized group member ID retrieval during broadcast
**Learning:** In Go/GORM applications broadcasting messages, querying full user objects just to extract IDs via joins is an unnecessary N+1-like performance bottleneck, especially in groups with many members.
**Action:** Used `Pluck` directly on `GroupMember` to fetch just `user_id`, saving bandwidth, processing, and memory allocation. Always consider if only an ID list is needed instead of the full model.
