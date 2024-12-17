package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func SplitString(input string) (int, int, error) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format: %s", input)
	}

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number: %v", err)
	}

	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number: %v", err)
	}

	return x, y, nil
}
