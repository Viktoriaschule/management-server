package reservation

import (
	"github.com/gin-gonic/gin"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/models"
)

func Serve(root *gin.RouterGroup, database *database.Database) {
	GroupServe(root, database)
	ListServe(root, database)

	// The combined overview for the current user
	root.GET("/", func(c *gin.Context) {
		username, _, _ := helper.GetAuth(c.Request)
		withHistory := c.DefaultQuery("history", "false") == "true"
		allUsers := c.DefaultQuery("allUsers", "true") == "true"

		participantFilter := "participant=\"" + username + "\""
		reservationFilter := ""
		if !allUsers {
			reservationFilter = participantFilter
		}

		reservations, err := getReservations(reservationFilter, database, !withHistory)
		groups, _err := getGroups(participantFilter, database)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if _err != nil {
			c.JSON(500, gin.H{"error": _err.Error()})
			return
		}

		c.JSON(200, gin.H{"reservations": reservations, "groups": groups})
		return
	})
}

func ListServe(root *gin.RouterGroup, database *database.Database) {
	root.POST("/list", func(c *gin.Context) {
		request := models.Reservation{}

		if err := c.ShouldBindJSON(&request); err == nil {
			request.Id = -1
			status, err := createOrUpdateReservation(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(status, gin.H{"reservation": request})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.PUT("/list", func(c *gin.Context) {
		request := models.Reservation{}

		if err := c.ShouldBindJSON(&request); err == nil {
			status, err := createOrUpdateReservation(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(status, gin.H{"reservation": request})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.DELETE("/list", func(c *gin.Context) {
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

func GroupServe(root *gin.RouterGroup, database *database.Database) {
	root.GET("/groups", func(c *gin.Context) {
		username, _, _ := helper.GetAuth(c.Request)
		reservations, err := getGroups("participant=\""+username+"\"", database)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"groups": reservations})
		return
	})

	root.POST("/groups", func(c *gin.Context) {
		request := models.Group{}

		if err := c.ShouldBindJSON(&request); err == nil {
			request.Id = -1
			err := createOrUpdateGroup(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"status": "Success", "group": request})
			return
		}

		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.PUT("/groups", func(c *gin.Context) {
		request := models.Group{}

		if err := c.ShouldBindJSON(&request); err == nil {
			err := createOrUpdateGroup(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"status": "Success", "group": request})
			return
		}

		c.JSON(400, gin.H{"error": "Wrong body format"})
	})

	root.DELETE("/groups", func(c *gin.Context) {
		request := models.Group{}

		if err := c.ShouldBindJSON(&request); err == nil {
			err := deleteGroup(&request, database)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "Success"})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})
}
