package reservation

import (
	"github.com/gin-gonic/gin"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/models"
)

func Serve(root *gin.RouterGroup, database *database.Database) {
	root.POST("/create", func(c *gin.Context) {
		request := models.Reservation{}

		if err := c.ShouldBindJSON(&request); err == nil {
			request.Id = -1
			err := createOrUpdateReservation(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "Success"})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.PUT("/update", func(c *gin.Context) {
		request := models.Reservation{}

		if err := c.ShouldBindJSON(&request); err == nil {
			err := createOrUpdateReservation(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "Success"})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.DELETE("/delete", func(c *gin.Context) {
		request := models.Reservation{}

		if err := c.ShouldBindJSON(&request); err == nil {
			err := deleteReservation(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "Success"})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.GET("/list", func(c *gin.Context) {
		withHistory := c.DefaultQuery("history", "false") == "true"
		reservations, err := getReservations("", database, !withHistory)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"reservations": reservations})
		return
	})

	root.GET("/list/:iPadGroup", func(c *gin.Context) {
		withHistory := c.DefaultQuery("history", "false") == "true"
		iPadGroup := c.Param("iPadGroup")
		reservations, err := getReservations(`ipad_group="`+iPadGroup+`"`, database, !withHistory)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"reservations": reservations})
		return
	})
}
