package database

// Sort represents a sorting configuration with a key and order.
type Sort struct {
	Key   string
	Order Order
}

// Order represents the sorting order (ascending or descending).
type Order int

const (
	OrderASC Order = iota
	OrderDESC
)
