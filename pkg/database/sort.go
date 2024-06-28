package database

// Sort represents a sorting configuration with a key and order.
type Sort struct {
	Key   string // Key represents the field to sort by.
	Order Order  // Order represents the sorting order (ascending or descending).
}

// Order represents the sorting order.
type Order int

const (
	OrderASC  Order = iota // OrderASC represents the ascending sorting order.
	OrderDESC              // OrderDESC represents the descending sorting order.
)
