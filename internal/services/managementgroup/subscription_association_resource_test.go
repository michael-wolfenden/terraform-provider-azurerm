package managementgroup_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-05-01/managementgroups"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/managementgroup/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ManagementGroupSubscriptionAssociation struct{}

// NOTE: this is a combined test rather than separate split out tests due to
// all testcases in this file share the same subscription instance so that
// these testcases have to be run sequentially.

func TestAccManagementGroupSubscriptionAssociation(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic":          testAccManagementGroupSubscriptionAssociation_basic,
			"requiresImport": testAccManagementGroupSubscriptionAssociation_requiresImport,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccManagementGroupSubscriptionAssociation_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group_subscription_association", "test")

	r := ManagementGroupSubscriptionAssociation{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func testAccManagementGroupSubscriptionAssociation_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group_subscription_association", "test")

	r := ManagementGroupSubscriptionAssociation{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r ManagementGroupSubscriptionAssociation) basic() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {
  subscription_id = %q
}

resource "azurerm_management_group" "test" {
}

resource "azurerm_management_group_subscription_association" "test" {
  management_group_id = azurerm_management_group.test.id
  subscription_id     = data.azurerm_subscription.test.id
}
`, os.Getenv("ARM_SUBSCRIPTION_ID_ALT"))
}

func (r ManagementGroupSubscriptionAssociation) requiresImport(_ acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_management_group_subscription_association" "import" {
  management_group_id = azurerm_management_group_subscription_association.test.management_group_id
  subscription_id     = azurerm_management_group_subscription_association.test.subscription_id
}
`, r.basic())
}

func (r ManagementGroupSubscriptionAssociation) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ManagementGroupSubscriptionAssociationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.ManagementGroups.GroupsClient.Get(ctx, id.ManagementGroup, "children", utils.Bool(false), "", "no-cache")
	if err != nil {
		return nil, fmt.Errorf("retrieving Management Group to check for Subscription Association: %+v", err)
	}

	if resp.Properties == nil || resp.Properties.Children == nil {
		return utils.Bool(false), nil
	}

	present := false
	for _, v := range *resp.Children {
		if v.Type == managementgroups.Type1Subscriptions && v.Name != nil && *v.Name == id.SubscriptionId {
			present = true
		}
	}

	return utils.Bool(present), nil
}
