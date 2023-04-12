package unbroken

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type GolangSuite struct {
	suite.Suite

	router *gin.Engine
}

const (
	hostAddr = "http://localhost:8080"
)

func (s *GolangSuite) SetupSuite() {
	g := gin.Default()
	RegisterGoHandlers(g)

	s.router = g
}

func (s *GolangSuite) TestGotestOutput() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("gotest", "go_test.out")
	s.NoError(err)

	file, err := os.Open("fixtures/go_test.out")
	s.NoError(err)

	_, err = io.Copy(fw, file)
	s.NoError(err)
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/go/push", hostAddr), bytes.NewReader(body.Bytes()))
	s.NoError(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	s.Equal(http.StatusOK, resp.Code)
}

func TestGolangSuite(t *testing.T) {
	suite.Run(t, new(GolangSuite))
}
