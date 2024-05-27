package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ianatha/websmsd/modem"
)

type OutgoingSMS struct {
	Number string `json:"number" binding:"required"`
	Msg    string `json:"msg" binding:"required"`
}

func main() {
	// 51  usb_modeswitch -v 0x12d1 -p 0x1f01 -V 0x12d1 -P 0x1001 -M "55534243000000000000000000000611060000000000000000000000000000"
	modem := modem.New("/dev/ttyUSB0", 115200)
	err := modem.Connect()
	if err != nil {
		panic(err)
	}
	router := gin.Default()

	router.GET("/sms", func(c *gin.Context) {
		smsList := modem.SMSReadAll("ALL")
		c.JSON(200, smsList)
	})

	router.POST("/sms", func(c *gin.Context) {
		var sms OutgoingSMS
		if c.ShouldBindJSON(&sms) == nil {
			err := modem.SMSSend(sms.Number, sms.Msg)
			if err != "OK\r\n" {
				c.JSON(500, gin.H{"status": "unable to send the message: " + err})
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
