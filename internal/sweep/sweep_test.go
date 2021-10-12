//go:build sweep
// +build sweep

package sweep_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	_ "github.com/hashicorp/terraform-provider-aws/internal/service/internal"
	_ "github.com/hashicorp/terraform-provider-aws/internal/service/service"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func TestMain(m *testing.M) {
	sweep.SweeperClients = make(map[string]interface{})
	resource.TestMain(m)
}
