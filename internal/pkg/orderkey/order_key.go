package orderkey

import "strconv"

const firstValue uint64 = 1
const fixedWidth = 12
const base = 36
const maxFixedWidthKey = "zzzzzzzzzzzz"
const minFixedWidthKey = "000000000000"

func First() string {
	return format(firstValue)
}

func After(prev string) string {
	if prev == "" {
		return First()
	}

	value, err := strconv.ParseUint(prev, base, 64)
	if err != nil {
		return prev + "0"
	}

	nextValue := value + 1
	if nextValue <= value {
		return prev + "0"
	}

	nextKey := format(nextValue)
	if len(prev) == fixedWidth && prev != maxFixedWidthKey {
		return nextKey
	}

	if nextKey > prev {
		return nextKey
	}

	return prev + "0"
}

func Before(next string) string {
	if next == "" {
		return First()
	}

	value, err := strconv.ParseUint(next, base, 64)
	if err != nil {
		return "-" + next
	}

	if value == 0 {
		return "-" + next
	}

	prevValue := value - 1

	prevKey := format(prevValue)
	if len(next) == fixedWidth && next != minFixedWidthKey {
		return prevKey
	}

	if prevKey < next {
		return prevKey
	}

	return "-" + next
}

func format(value uint64) string {
	key := strconv.FormatUint(value, base)
	for len(key) < fixedWidth {
		key = "0" + key
	}

	return key
}
