package rest

import (
	"fmt"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/history"
	"github.com/viktoriaschule/management-server/reservation"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/viktoriaschule/management-server/auth"
	"github.com/viktoriaschule/management-server/config"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/relution"
)

func Serve(config *config.Config, database *database.Database) {
	r := gin.New()
	r.Use(gin.Recovery())
	if log.Level >= log.Debug {
		r.Use(gin.Logger())
	}

	root := r.Group("/", basicAuth(config))

	relution.Serve(root.Group("/ipad_list"), database)
	history.Serve(root.Group("/history"), database)
	reservation.Serve(root.Group("/reservations"), database)

	err := r.Run(fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Errorf("Error serving API: %v", err)
		os.Exit(1)
	}
}

func basicAuth(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, status := helper.GetAuth(c.Request)

		if status == 401 || !authenticateUser(username, password, config) {
			c.Writer.Header().Set("WWW-Authenticate", "Basic")
			respondWithError(401, "Unauthorized", c)
			return
		}

		c.Next()
	}
}

func authenticateUser(username, password string, config *config.Config) bool {
	result, err := auth.CheckUser(username, password, config)
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
