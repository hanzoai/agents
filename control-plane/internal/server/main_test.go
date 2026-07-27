package server

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain sets gin's mode once for the whole package. Each test calling
// gin.SetMode itself races under -race, because SetMode writes package-level
// state and the tests run in parallel.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
