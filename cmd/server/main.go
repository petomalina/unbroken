package main

import (
	"github.com/gin-gonic/gin"
	"github.com/petomalina/unbroken"
)

func main() {
	r := gin.Default()

	unbroken.RegisterGoHandlers(r)

	r.Run() // listen and serve on 0.0.0.0:8080
}
