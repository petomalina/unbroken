package unbroken

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"strconv"
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
			// skip empty tests, as these outputs sometimes happen due to framework configuration
			if goTestLine.Test == "" {
				continue
			}

			// skip suites as they are not tests
			if strings.HasSuffix(goTestLine.Test, "Suite") {
				continue
			}

			log.Info("test results",
				zap.String("package", goTestLine.Package),
				zap.String("test", goTestLine.Test),
				zap.String("status", goTestLine.Action),
			)
		}

		if goTestLine.Action == "output" {
			// only lines starting with ok seem to be the coverage results
			if !strings.HasPrefix(goTestLine.Output, "ok") || !strings.Contains(goTestLine.Output, "coverage:") {
				continue
			}

			coverageSplit := strings.Split(goTestLine.Output, "coverage: ")
			if len(coverageSplit) != 2 {
				log.Error("error parsing coverage", zap.Any("line", goTestLine.Output))
				continue
			}

			percentSplit := strings.Split(coverageSplit[1], "%")
			if len(percentSplit) != 2 {
				log.Error("error parsing coverage", zap.Any("line", goTestLine.Output))
				continue
			}

			coverage, err := strconv.ParseFloat(percentSplit[0], 64)
			if err != nil {
				log.Error("error parsing coverage", zap.Error(err))
				continue
			}

			log.Info("coverage results",
				zap.String("package", goTestLine.Package),
				zap.Float64("coverage", coverage),
			)
		}
	}

	return nil
}
