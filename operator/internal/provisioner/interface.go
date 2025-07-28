/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provisioner

import (
	"context"

	dbtreev1 "github.com/piper-hyowon/dBtree/operator/api/v1"
)

// Provisioner interface defines methods for provisioning database instances
type Provisioner interface {
	// Provision creates all necessary resources for the database instance
	Provision(ctx context.Context, instance *dbtreev1.DBInstance) error

	// Delete removes all resources associated with the database instance
	Delete(ctx context.Context, instance *dbtreev1.DBInstance) error

	// Update modifies existing resources based on spec changes
	Update(ctx context.Context, instance *dbtreev1.DBInstance) error

	// GetStatus retrieves the current status of the database instance
	GetStatus(ctx context.Context, instance *dbtreev1.DBInstance) (*dbtreev1.DBInstanceStatus, error)
}
