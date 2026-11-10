package get

// Display fields historically treat an omitted number as zero. The machine
// projection retains the original pointer to distinguish absence from zero.
func inventoryNumber[T int | int64](reported *T) T {
	if reported == nil {
		return 0
	}
	return *reported
}
