package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSecurityGroup_importBasic(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 3: group, 2 rules
		if len(s) != 3 {
			return fmt.Errorf("expected 3 states: %#v", s)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfig,
			},

			resource.TestStep{
				ResourceName:     "aws_security_group.web",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}
