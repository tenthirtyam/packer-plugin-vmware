// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Step to discover the http ip
// which guests use to reach the vm host
// To make sure the IP is set before boot command and http server steps
type StepHTTPIPDiscover struct{}

func (s *StepHTTPIPDiscover) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// Determine the host IP
	hostIP, err := driver.HostIP(state)
	if err != nil {
		err = fmt.Errorf("error detecting host IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[INFO] Host IP for the virtual machine: %s", hostIP)
	state.Put("http_ip", hostIP)

	return multistep.ActionContinue
}

func (*StepHTTPIPDiscover) Cleanup(multistep.StateBag) {}
