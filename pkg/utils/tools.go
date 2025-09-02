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

package utils

import (
	"fmt"
	"strconv"
)

func FormatCpuCapacity(cpuHz int64) string {
	cpuGHz := float64(cpuHz) / 1000000000
	return strconv.FormatFloat(cpuGHz, 'f', 2, 64) + " GHz"
}

func FormatMemorySize(sizeInBytes int64) string {
	if sizeInBytes < 1024 {
		return fmt.Sprintf("%d B", sizeInBytes)
	} else if sizeInBytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(sizeInBytes)/1024)
	} else if sizeInBytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(sizeInBytes)/(1024*1024))
	} else if sizeInBytes < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(sizeInBytes)/(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2f TB", float64(sizeInBytes)/(1024*1024*1024*1024))
	}
}

func FormatDiskSize(sizeInBytes int64) string {
	return FormatMemorySize(sizeInBytes)
}
