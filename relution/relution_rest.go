package relution

import (
	"github.com/gin-gonic/gin"
	"github.com/viktoriaschule/management-server/database"
)

func Serve(root *gin.RouterGroup, database *database.Database) {
	root.GET("/devices", func(c *gin.Context) {
		devices, err := GetValidLoadedDevices(database)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"devices": devices})
	})
}
