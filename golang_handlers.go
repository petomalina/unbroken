package unbroken

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterGoHandlers(g *gin.Engine) {
	group := g.Group("/go")
	group.POST("/push", handleGoPush)
}

func handleGoPush(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	gotestFile, ok := form.File["gotest"]
	if ok {
		for _, f := range gotestFile {
			err := parseGoTestFile(f)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
		}
	}
}

func parseGoTestFile(file *multipart.FileHeader) error {
	log.Info("parsing gotest file", zap.Any("file", file.Filename))

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	bb, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bb), "\n")
	for _, line := range lines {
		goTestLine := &GoTestLine{}
		err := json.Unmarshal([]byte(line), goTestLine)
		if err != nil {
			log.Error("error parsing json", zap.Error(err))
			return err
		}

		if goTestLine.Action == "pass" || goTestLine.Action == "fail" {
			log.Info("gotestline", zap.Any("line", goTestLine))
		}
	}

	return nil
}
