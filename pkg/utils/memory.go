// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// pkg/utils/memory.go
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParseMemorySize(memoryStr string) (int64, error) {
	if memoryStr == "" {
		return 0, fmt.Errorf("memory size cannot be empty")
	}

	memoryStr = strings.TrimSpace(strings.ToUpper(memoryStr))

	re := regexp.MustCompile(`^(\d+)([KMGT]?B?)?$`)
	matches := re.FindStringSubmatch(memoryStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid memory format: %s. Use format like 1G, 1024M", memoryStr)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory size: %v", err)
	}

	unit := matches[2]
	if unit == "" || unit == "B" {
		return value, nil
	}

	switch unit[0] {
	case 'K':
		return value * 1024, nil
	case 'M':
		return value * 1024 * 1024, nil
	case 'G':
		return value * 1024 * 1024 * 1024, nil
	case 'T':
		return value * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory unit: %s", unit)
	}
}
