package unbroken

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"os"
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

	metrics := []*PrometheusMetric{}

	gotestFile, ok := form.File["gotest"]
	if ok {
		for _, f := range gotestFile {
			mm, err := parseGoTestFile(f)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			metrics = append(metrics, mm...)
		}
	}

	if os.Getenv("UNBROKEN_PUSH_METRICS") == "true" {
		grafanaURL := os.Getenv("UNBROKEN_GRAFANA_URL")
		grafanaKey := os.Getenv("UNBROKEN_GRAFANA_KEY")

		err := PushToGrafana(metrics, grafanaURL, grafanaKey)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
	}
}

func parseGoTestFile(file *multipart.FileHeader) ([]*PrometheusMetric, error) {
	log.Info("parsing gotest file", zap.Any("file", file.Filename))

	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	metrics := []*PrometheusMetric{}

	lines := strings.Split(string(bb), "\n")
	for _, line := range lines {
		goTestLine := &GoTestLine{}
		err := json.Unmarshal([]byte(line), goTestLine)
		if err != nil {
			log.Error("error parsing json", zap.Error(err))
			return nil, err
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

			var suite, test string
			if strings.Contains(goTestLine.Test, "/") {
				split := strings.Split(goTestLine.Test, "/")
				if len(split) != 2 {
					log.Error("error parsing test", zap.Any("line", goTestLine.Test))
					continue
				}
				suite = split[0]
				test = split[1]
			} else {
				suite = ""
				test = goTestLine.Test
			}

			// we must convert the test result to an int, as prometheus does not support strings via this method
			value := "1"
			if goTestLine.Action == "fail" {
				value = "0"
			}

			metrics = append(metrics, &PrometheusMetric{
				Name:  "go_test",
				Value: value,
				Labels: map[string]string{
					"package": strings.Replace(strings.Replace(goTestLine.Package, "/", "_", -1), ".", "_", -1),
					"suite":   suite,
					"test":    test,
				},
			})
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
			coverageString := percentSplit[0]

			// coverage, err := strconv.ParseFloat(coverageString, 64)
			// if err != nil {
			// 	log.Error("error parsing coverage", zap.Error(err))
			// 	continue
			// }

			metrics = append(metrics, &PrometheusMetric{
				Name:  "go_coverage",
				Value: coverageString,
				Labels: map[string]string{
					"package": goTestLine.Package,
				},
			})
		}
	}

	return metrics, nil
}
