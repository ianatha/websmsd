package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/xlab/at"
)

func TimeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(2500*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(func(c *gin.Context) {
			c.String(http.StatusRequestTimeout, "timeout")
		}),
	)
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		log.Printf("Error: %s\n", err.Error())
	}

	c.JSON(http.StatusInternalServerError, "")
}

func (m *Monitor) newGinEngine() *gin.Engine {
	router := gin.Default()
	router.Use(ErrorHandler)
	router.Use(TimeoutMiddleware())

	router.GET("/", func(c *gin.Context) {
		data := struct {
			Mon  *Monitor
			Dev  *at.Device
			Time time.Time
		}{
			Mon:  m,
			Dev:  m.dev,
			Time: time.Now(),
		}

		var buf bytes.Buffer
		err := tpl.Execute(&buf, data)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})

	router.GET("/stats", func(c *gin.Context) {
		c.JSON(200, m.DeviceState())
	})

	router.GET("/sms", func(c *gin.Context) {
		c.JSON(200, m.Messages)
	})

	// router.POST("/sms", func(c *gin.Context) {
	// 	var sms OutgoingSMS
	// 	if c.ShouldBindJSON(&sms) == nil {
	// 		err := modem.SMSSend(sms.Number, sms.Msg)
	// 		if err != nil {
	// 			c.AbortWithError(405, err)
	// 			return
	// 		}
	// 		c.JSON(200, gin.H{"status": "Message sent"})
	// 		return
	// 	} else {
	// 		c.JSON(400, gin.H{"status": "Invalid request body"})
	// 		return
	// 	}
	// })

	router.DELETE("/sms/:id", func(c *gin.Context) {
		id := c.Param("id")
		m.deleteMessageWithId(id)
		c.JSON(200, gin.H{
			"status": "deleted",
		})
	})

	return router
}
