package middleware

import (
	"compress/gzip"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UseGzip(c *gin.Context) {
	// Decompress when Content-Encoding announces gzip; otherwise pass the
	// body through untouched. The pagehide / keepalive code path in the SDK
	// has to dispatch the request synchronously inside the unload handler,
	// which means it can't `await` the async CompressionStream — those
	// requests arrive as plain JSON and we accept them as-is.
	if c.GetHeader("Content-Encoding") != "gzip" {
		c.Next()
		return
	}

	gzReader, err := gzip.NewReader(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip"})
		return
	}
	c.Request.Body = gzReader
	c.Next()
}
