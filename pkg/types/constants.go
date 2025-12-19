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

// Package types provides shared constants and type definitions for the CLI.
// This package has no dependencies on other CLI packages to avoid import cycles.
package types

// VM Instance States
const (
	VMStateRunning   = "Running"
	VMStateStopped   = "Stopped"
	VMStatePaused    = "Paused"
	VMStateDestroyed = "Destroyed"
	VMStateCreating  = "Creating"
	VMStateStarting  = "Starting"
	VMStateStopping  = "Stopping"
	VMStateRebooting = "Rebooting"
	VMStateMigrating = "Migrating"
)

// Image States
const (
	ImageStateEnabled  = "Enabled"
	ImageStateDisabled = "Disabled"
)

// Image Status
const (
	ImageStatusReady       = "Ready"
	ImageStatusDownloading = "Downloading"
	ImageStatusDeleted     = "Deleted"
)

// Resource States (generic)
const (
	StateEnabled  = "Enabled"
	StateDisabled = "Disabled"
)

// IsVMRunnable checks if a VM is in a state that can be started
func IsVMRunnable(state string) bool {
	return state == VMStateStopped
}

// IsVMStoppable checks if a VM is in a state that can be stopped
func IsVMStoppable(state string) bool {
	return state == VMStateRunning || state == VMStatePaused
}

// IsVMActive checks if a VM is in an active state (not destroyed)
func IsVMActive(state string) bool {
	return state == VMStateRunning || state == VMStateStopped || state == VMStatePaused
}

// IsImageReady checks if an image is ready for use
func IsImageReady(status string) bool {
	return status == ImageStatusReady
}
