// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown

package common

import (
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Set the firmware types for the virtual machine.
const (
	FirmwareTypeBios       = "bios"
	FirmwareTypeUEFI       = "efi"
	FirmwareTypeUEFISecure = "efi-secure"
)

// allowedFirmwareTypes is a list of allowed firmware types for the virtual
// machine.
var allowedFirmwareTypes = []string{FirmwareTypeBios, FirmwareTypeUEFI, FirmwareTypeUEFISecure}

type HWConfig struct {
	// The firmware type for the virtual machine.
	// Allowed values are `bios`, `efi`, and `efi-secure` (for secure boot).
	// Defaults to the recommended firmware type for the guest operating system.
	Firmware string `mapstructure:"firmware" required:"false"`
	// The number of virtual CPUs cores for the virtual machine.
	CpuCount int `mapstructure:"cpus" required:"false"`
	// The number of virtual CPU cores per socket for the virtual machine.
	CoreCount int `mapstructure:"cores" required:"false"`
	// The amount of memory for the virtual machine in MB.
	MemorySize int `mapstructure:"memory" required:"false"`
	// The network which the virtual machine will connect for local desktop
	// hypervisors. Use the generic values that map to a device, such as
	// `hostonly`, `nat`, or `bridged`. Defaults to `nat`.
	//
	// ~> **Note:** If not set to one of these generic values, then it is
	// assumed to be a network device (_e.g._, `VMnet0..x`).
	Network string `mapstructure:"network" required:"false"`
	// The network which the virtual machine will connect on a remote
	// hypervisor.
	NetworkName string `mapstructure:"network_name" required:"false"`
	// The virtual machine network card type. Recommended values are `e1000` and
	// `vmxnet3`. Defaults to `e1000`.
	//
	// Refer to VMware product documentation for supported network adapter types
	// for the hypervisor and guest operating system.
	NetworkAdapterType string `mapstructure:"network_adapter_type" required:"false"`
	// Enable virtual sound card device. Defaults to `false`.
	Sound bool `mapstructure:"sound" required:"false"`
	// Enable USB 2.0 controllers for the virtual machine.
	// Defaults to `false`.
	//
	// ~> **Note:** To enable USB 3.0 controllers, set a `usb_xhci.present`
	// key to `true` in the `vmx_data` option.
	USB bool `mapstructure:"usb" required:"false"`
	// Add a serial port to the virtual machine. Use a format of
	// `Type:option1,option2,...`. Allowed values for the field `Type` include:
	// `FILE`, `DEVICE`, `PIPE`, `AUTO`, or `NONE`.
	//
	// * `FILE:path(,yield)` - Specifies the path to the local file to be used
	//   as the serial port.
	//
	//   * `yield` (bool) - This is an optional boolean that specifies
	//     whether the virtual machine should yield the CPU when polling the
	//     port. By default, the builder will assume this as `FALSE`.
	//
	// * `DEVICE:path(,yield)` - Specifies the path to the local device to be
	//   used as the serial port. If `path` is empty, then default to the first
	//   serial port.
	//
	//   * `yield` (bool) - This is an optional boolean that specifies
	//     whether the virtual machine should yield the CPU when polling the
	//     port. By default, the builder will assume this as `FALSE`.
	//
	// * `PIPE:path,endpoint,host(,yield)` - Specifies to use the named-pipe
	//   "path" as a serial port. This has a few options that determine how the
	//   VM should use the named-pipe.
	//
	//   * `endpoint` (string) - Chooses the type of the VM-end, which can be
	//     either a `client` or `server`.
	//
	//   * `host` (string) - Chooses the type of the host-end, which can be
	//     either `app` (application) or `vm` (another virtual-machine).
	//
	//   * `yield` (bool) - This is an optional boolean that specifies whether
	//     the virtual machine should yield the CPU when polling the port. By
	//     default, the builder will assume this as `FALSE`.
	//
	// * `AUTO: (yield)` - Specifies to use auto-detection to determine the
	//   serial port to use. This has one option to determine how the virtual
	//   machine should support the serial port.
	//
	//   * `yield` (bool) - This is an optional boolean that specifies whether
	//     the virtual machine should yield the CPU when polling the port. By
	//     default, the builder will assume this as `FALSE`.
	//
	// * `NONE` - Specifies to not use a serial port. (default)
	Serial string `mapstructure:"serial" required:"false"`
	// Add a parallel port to add to the virtual machine. Use a format of
	// `Type:option1,option2,...`. Allowed values for the field `Type` include:
	// `FILE`, `DEVICE`, `AUTO`, or `NONE`.
	//
	// * `FILE:path` - Specifies the path to the local file to be used for the
	//    parallel port.
	//
	// * `DEVICE:path` - Specifies the path to the local device to be used for
	//    the parallel port.
	//
	// * `AUTO:direction` - Specifies to use auto-detection to determine the
	//   parallel port. Direction can be `BI` to specify bidirectional
	//   communication or `UNI` to specify unidirectional communication.
	//
	// * `NONE` - Specifies to not use a parallel port. (default)
	Parallel string `mapstructure:"parallel" required:"false"`
}

func (c *HWConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if (c.Firmware != "") && (!slices.Contains(allowedFirmwareTypes, c.Firmware)) {
		errs = append(errs, fmt.Errorf("invalid 'firmware' type specified: %s; must be one of %s", c.Firmware, strings.Join(allowedFirmwareTypes, ", ")))
	}

	if c.CpuCount < 0 {
		errs = append(errs, fmt.Errorf("invalid number of cpus specified (cpus < 0): %d", c.CpuCount))
	}

	if c.CoreCount < 0 {
		errs = append(errs, fmt.Errorf("invalid number of cpu cores specified (cores < 0): %d", c.CoreCount))
	}

	if c.MemorySize < 0 {
		errs = append(errs, fmt.Errorf("invalid amount of memory specified (memory < 0): %d", c.MemorySize))
	}

	// Peripherals
	if !c.Sound {
		c.Sound = false
	}

	if !c.USB {
		c.USB = false
	}

	if c.Parallel == "" {
		c.Parallel = "none"
	}

	if c.Serial == "" {
		c.Serial = "none"
	}

	return errs
}

type ParallelUnion struct {
	Union  interface{}
	File   *ParallelPortFile
	Device *ParallelPortDevice
	Auto   *ParallelPortAuto
}

type ParallelPortFile struct {
	Filename string
}

type ParallelPortDevice struct {
	Bidirectional string
	Devicename    string
}

type ParallelPortAuto struct {
	Bidirectional string
}

func (c *HWConfig) HasParallel() bool {
	return c.Parallel != ""
}

func (c *HWConfig) ReadParallel() (*ParallelUnion, error) {
	input := strings.SplitN(c.Parallel, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("unexpected format for parallel port: %s", c.Parallel)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "FILE":
		res := &ParallelPortFile{Filename: filepath.FromSlash(formatOptions)}
		return &ParallelUnion{Union: res, File: res}, nil
	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 1 || len(comp) > 2 {
			return nil, fmt.Errorf("unexpected format for parallel port: %s", c.Parallel)
		}
		res := new(ParallelPortDevice)
		res.Bidirectional = "FALSE"
		res.Devicename = filepath.FromSlash(comp[0])
		if len(comp) > 1 {
			switch strings.ToUpper(comp[1]) {
			case "BI":
				res.Bidirectional = "TRUE"
			case "UNI":
				res.Bidirectional = "FALSE"
			default:
				return nil, fmt.Errorf("unknown direction %s specified for parallel port: %s", strings.ToUpper(comp[1]), c.Parallel)
			}
		}
		return &ParallelUnion{Union: res, Device: res}, nil

	case "AUTO":
		res := new(ParallelPortAuto)
		switch strings.ToUpper(formatOptions) {
		case "":
			fallthrough
		case "UNI":
			res.Bidirectional = "FALSE"
		case "BI":
			res.Bidirectional = "TRUE"
		default:
			return nil, fmt.Errorf("unknown direction %s specified for parallel port: %s", strings.ToUpper(formatOptions), c.Parallel)
		}
		return &ParallelUnion{Union: res, Auto: res}, nil

	case "NONE":
		return &ParallelUnion{Union: nil}, nil
	}

	return nil, fmt.Errorf("unexpected format for parallel port: %s", c.Parallel)
}

type SerialConfigPipe struct {
	Filename string
	Endpoint string
	Host     string
	Yield    string
}

type SerialConfigFile struct {
	Filename string
	Yield    string
}

type SerialConfigDevice struct {
	Devicename string
	Yield      string
}

type SerialConfigAuto struct {
	Devicename string
	Yield      string
}

type SerialUnion struct {
	Union  interface{}
	Pipe   *SerialConfigPipe
	File   *SerialConfigFile
	Device *SerialConfigDevice
	Auto   *SerialConfigAuto
}

func (c *HWConfig) HasSerial() bool {
	return c.Serial != ""
}

func (c *HWConfig) ReadSerial() (*SerialUnion, error) {
	var defaultSerialPort string
	if runtime.GOOS == "windows" {
		defaultSerialPort = "COM1"
	} else {
		defaultSerialPort = "/dev/ttyS0"
	}
	input := strings.SplitN(c.Serial, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("unexpected format for serial port: %s", c.Serial)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "PIPE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 3 || len(comp) > 4 {
			return nil, fmt.Errorf("unexpected format for serial port pipe: %s", c.Serial)
		}
		if res := strings.ToLower(comp[1]); res != "client" && res != "server" {
			return nil, fmt.Errorf("unexpected format for endpoint in serial port pipe: %s -> %s", c.Serial, res)
		}
		if res := strings.ToLower(comp[2]); res != "app" && res != "vm" {
			return nil, fmt.Errorf("unexpected format for host in serial port pipe: %s -> %s", c.Serial, res)
		}
		res := &SerialConfigPipe{
			Filename: comp[0],
			Endpoint: comp[1],
			Host:     map[string]string{"app": "TRUE", "vm": "FALSE"}[strings.ToLower(comp[2])],
			Yield:    "FALSE",
		}
		if len(comp) == 4 {
			res.Yield = strings.ToUpper(comp[3])
		}
		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("unexpected format for yield in serial port pipe: %s -> %s", c.Serial, res.Yield)
		}
		return &SerialUnion{Union: res, Pipe: res}, nil

	case "FILE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("unexpected format for serial port file: %s", c.Serial)
		}

		res := &SerialConfigFile{Yield: "FALSE"}
		res.Filename = filepath.FromSlash(comp[0])
		res.Yield = "FALSE"
		if len(comp) > 1 {
			res.Yield = strings.ToUpper(comp[1])
		}
		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("unexpected format for yield in serial port file: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, File: res}, nil

	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("unexpected format for serial port device: %s", c.Serial)
		}
		res := new(SerialConfigDevice)
		// set serial port defaults
		res.Devicename = defaultSerialPort
		res.Yield = "FALSE"
		// Read actual values from component, if set.
		if len(comp) == 2 {
			res.Devicename = filepath.FromSlash(comp[0])
			res.Yield = strings.ToUpper(comp[1])
		}

		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("unexpected format for yield in serial port device: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, Device: res}, nil

	case "AUTO":
		res := new(SerialConfigAuto)
		res.Devicename = defaultSerialPort

		if len(formatOptions) > 0 {
			res.Yield = strings.ToUpper(formatOptions)
		} else {
			res.Yield = "FALSE"
		}

		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("unexpected format for yield in serial port auto: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, Auto: res}, nil

	case "NONE":
		return &SerialUnion{Union: nil}, nil

	default:
		return nil, fmt.Errorf("unknown serial type %s: %s", strings.ToUpper(formatType), c.Serial)
	}
}
