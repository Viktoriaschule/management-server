package rest

import (
	"encoding/base64"
	"fmt"
	"github.com/viktoriaschule/management-server/relution"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/viktoriaschule/management-server/auth"
	"github.com/viktoriaschule/management-server/config"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/log"
)

func Serve(config *config.Config, database *database.Database) {
	r := gin.Default()
	root := r.Group("/", basicAuth())
	root.GET("/devices", func(c *gin.Context) {
		devices, err := relution.GetValidLoadedDevices(database)
		if err != nil {
			respondWithError(500, err.Error(), c)
		}
		c.JSON(200, gin.H{"devices": devices})
	})
	err := r.Run(fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Errorf("Error serving API: %v", err)
		os.Exit(1)
	}
}

func basicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			c.Writer.Header().Set("WWW-Authenticate", "Basic")
			respondWithError(401, "Unauthorized", c)
			return
		}
		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !authenticateUser(pair[0], pair[1]) {
			c.Writer.Header().Set("WWW-Authenticate", "Basic")
			respondWithError(401, "Unauthorized", c)
			return
		}

		c.Next()
	}
}

func authenticateUser(username, password string) bool {
	result, err := auth.CheckUser(username, password)
	if err != nil {
		log.Errorf("%v", err)
	}
	return result
}

func respondWithError(code int, message string, c *gin.Context) {
	resp := map[string]interface{}{
		"error": message,
	}

	c.JSON(code, resp)
	c.Abort()
}
