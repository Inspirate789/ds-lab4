package repository

const (
	// WARNING: when OFFSET is at least as great as the number of rows returned from the base query, no rows are returned.
	// So we get no full_count, either. If that's a rare case, just run a second query for the count in this case.
	selectRentalsQuery = `select *, count(*) over () as total_count from rentals where username = $3 offset $1 limit $2;`
	selectRentalQuery  = `select * from rentals where rental_uid = $1 limit 1;`
	insertRentalQuery  = `
		insert into rentals(rental_uid, username, payment_uid, car_uid, date_from, date_to, status) 
		values (:rental_uid, :username, :payment_uid, :car_uid, :date_from, :date_to, :status) 
		returning *;
	`
	updateRentalStatusQuery = `update rentals set status = $2 where rental_uid = $1;`
)
