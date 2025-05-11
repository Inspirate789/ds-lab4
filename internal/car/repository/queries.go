package repository

const (
	// WARNING: when OFFSET is at least as great as the number of rows returned from the base query, no rows are returned.
	// So we get no full_count, either. If that's a rare case, just run a second query for the count in this case.
	selectCarsQuery = `select *, count(*) over () as total_count from cars where $3 = true or availability = true offset $1 limit $2;`
	selectCarQuery  = `select * from cars where car_uid = $1 limit 1;`
	lockCarQuery    = `update cars set availability = false where car_uid = $1 and availability = true returning *;`
	unlockCarQuery  = `update cars set availability = true where car_uid = $1;`
)
