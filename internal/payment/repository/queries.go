package repository

const (
	// WARNING: when OFFSET is at least as great as the number of rows returned from the base query, no rows are returned.
	// So we get no full_count, either. If that's a rare case, just run a second query for the count in this case.
	insertPaymentQuery       = `insert into payments(payment_uid, status, price) values (:payment_uid, :status, :price) returning *;`
	selectPaymentQuery       = `select * from payments where payment_uid = $1 limit 1;`
	updatePaymentStatusQuery = `update payments set status = $2 where payment_uid = $1;`
)
