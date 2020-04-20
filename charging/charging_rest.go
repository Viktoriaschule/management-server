package charging

import (
	"github.com/gin-gonic/gin"

	"github.com/viktoriaschule/management-server/database"
)

func Serve(root *gin.RouterGroup, database *database.Database) {
	root.POST("/batteries", func(c *gin.Context) {
		request := BatteriesRequest{}

		if err := c.ShouldBindJSON(&request); err == nil {
			entries, err := GetBatteryEntriesForDevices(database, request.Ids, request.Date)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"devices": entries})
			return
		}
		c.JSON(400, gin.H{"error": "Wrong body format"})
	})
}
