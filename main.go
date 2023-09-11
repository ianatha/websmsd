package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/google/gousb"
	"github.com/ianatha/websmsd/modem"
)

type OutgoingSMS struct {
	Number string `json:"number" binding:"required"`
	Msg    string `json:"msg" binding:"required"`
}

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

// CheckUSBDeviceType checks the type of USB device
func CheckUSBDeviceType() USBDeviceType {
	ctx := gousb.NewContext()
	defer ctx.Close()

	USBIDs := map[string]USBDeviceType{
		"12d1:1f01": STORAGE,
		"12d1:1001": MODEM,
	}

	defer ctx.Close()

	var result USBDeviceType = NOT_FOUND
	_, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		log.Printf("Vendor: %s, Product: %s\n", desc.Vendor.String(), desc.Product.String())
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


func testResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "timeout")
}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(500*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(testResponse),
	)
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		log.Printf("Error: %s\n", err.Error())
	}

	c.JSON(http.StatusInternalServerError, "")
}

func main() {
	// Check if the device is a modem
	if CheckUSBDeviceType() == STORAGE {
		log.Printf("Switching to modem mode\n")
		runUSBModeSwitch()
	}

	if CheckUSBDeviceType() != MODEM {
		log.Fatal("No modem found")
		return
	}
	
	log.Printf("Modem found\n")
	modem := modem.New("/dev/ttyUSB0", 115200)
	err := modem.Connect()
	if err != nil {
		panic(err)
	}
	router := gin.Default()
	router.Use(ErrorHandler)
	router.Use(timeoutMiddleware())
	
	router.GET("/stats", func(c *gin.Context) {
		stats := modem.Stats(c)
		c.JSON(200, stats)
	})

	router.GET("/sms", func(c *gin.Context) {
		smsList, err := modem.SMSReadAll(c, "ALL")
		if err != nil {
			c.AbortWithError(405, err)
			return
		}
		c.JSON(200, smsList)
	})

	router.POST("/sms", func(c *gin.Context) {
		var sms OutgoingSMS
		if c.ShouldBindJSON(&sms) == nil {
			err := modem.SMSSend(sms.Number, sms.Msg)
			if err != nil {
				c.AbortWithError(405, err)
				return
			}
			c.JSON(200, gin.H{"status": "Message sent"})
			return
		} else {
			c.JSON(400, gin.H{"status": "Invalid request body"})
			return
		}
	})

	router.DELETE("/sms/:index", func(c *gin.Context) {
		indexStr := c.Param("index")
		indexInt, err := strconv.Atoi(indexStr)
		if err != nil {
			c.JSON(400, gin.H{
				"status": fmt.Sprintf("Invalid index %s", indexStr),
			})
			return
		}

		modem.SMSDelete(indexInt)
		c.JSON(200, gin.H{
			"status": "Message deleted",
		})
	})

	router.Run()
}
