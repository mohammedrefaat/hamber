package tools

func bitmapSort(arr []int64) []int64 {
	if len(arr) == 0 {
		return arr // Handle empty array
	}

	// Step 1: Find the min and max values in the array
	minVal, maxVal := arr[0], arr[0]
	for _, num := range arr {
		if num < minVal {
			minVal = num
		}
		if num > maxVal {
			maxVal = num
		}
	}

	// Step 2: Create a bitmap based on the detected range
	bitmapSize := maxVal - minVal + 1
	bitmap := make([]bool, bitmapSize)

	// Step 3: Populate the bitmap with the presence of each number
	for _, num := range arr {
		index := num - minVal // Normalize the numbers based on the minimum value
		bitmap[index] = true
	}

	// Step 4: Collect sorted numbers from the bitmap
	var sortedArr []int64
	for i, exists := range bitmap {
		if exists {
			sortedArr = append(sortedArr, minVal+int64(i))
		}
	}

	return sortedArr
}

func bitmapSortWithLimits(Numbers []int64, minVal, maxVal int64) []int64 {
	// Define the size of the bitmap based on the range of phone numbers
	bitmapSize := maxVal - minVal + 1
	bitmap := make([]bool, bitmapSize)

	// Mark presence of phone numbers in the bitmap
	for _, num := range Numbers {
		index := num - minVal // normalize to start from 0
		bitmap[index] = true
	}

	// Collect sorted phone numbers
	var sortedPhoneNumbers []int64
	for i, exists := range bitmap {
		if exists {
			sortedPhoneNumbers = append(sortedPhoneNumbers, minVal+int64(i))
		}
	}

	return sortedPhoneNumbers
}
