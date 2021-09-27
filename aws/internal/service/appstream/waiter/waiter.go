package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// stackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	stackOperationTimeout = 4 * time.Minute

	// fleetStateTimeout Maximum amount of time to wait for the statusFleetState to be RUNNING or STOPPED
	fleetStateTimeout = 180 * time.Minute
	// fleetOperationTimeout Maximum amount of time to wait for Fleet operation eventual consistency
	fleetOperationTimeout = 15 * time.Minute
)

// waitStackStateDeleted waits for a deleted stack
func waitStackStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"NotFound", "Unknown"},
		Refresh: statusStackState(ctx, conn, name),
		Timeout: stackOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Stack); ok {
		return output, err
	}

	return nil, err
}

// waitFleetStateRunning waits for a fleet running
func waitFleetStateRunning(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.FleetStateStarting},
		Target:  []string{appstream.FleetStateRunning},
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		return output, err
	}

	return nil, err
}

// waitFleetStateStopped waits for a fleet stopped
func waitFleetStateStopped(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.FleetStateStopping},
		Target:  []string{appstream.FleetStateStopped},
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		return output, err
	}

	return nil, err
}
