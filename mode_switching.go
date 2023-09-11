package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/google/gousb"
)

type USBDeviceType int

const (
	NOT_FOUND USBDeviceType = iota
	STORAGE
	MODEM
)

func (d USBDeviceType) String() string {
	return [...]string{"NOT_FOUND", "STORAGE", "MODEM"}[d]
}

func runUSBModeSwitch() error {
	cmd := exec.Command("usb_modeswitch", "-v", "0x12d1", "-p", "0x1f01", "-V", "0x12d1", "-P", "0x1001", "-M", "55534243000000000000000000000611060000000000000000000000000000")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func checkUSBDeviceType() USBDeviceType {
	ctx := gousb.NewContext()
	defer ctx.Close()

	USBIDs := map[string]USBDeviceType{
		"12d1:1f01": STORAGE,
		"12d1:1001": MODEM,
	}

	defer ctx.Close()

	var result USBDeviceType = NOT_FOUND
	_, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		log.Printf("found USB device, vendor=%s, product=%s\n", desc.Vendor.String(), desc.Product.String())
		if deviceType, ok := USBIDs[fmt.Sprintf("%s:%s", desc.Vendor.String(), desc.Product.String())]; ok {
			result = deviceType
			return false
		}
		return false
	})

	if err != nil {
		log.Fatalf("list: %s", err)
	}

	return result
}

func AssertHuaweiModemMode() {
	if checkUSBDeviceType() == STORAGE {
		log.Println("detected Huawei modem in STORAGE mode; switching to MODEM mode")
		runUSBModeSwitch()
	}
}
