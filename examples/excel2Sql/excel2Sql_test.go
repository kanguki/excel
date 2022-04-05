package excel2sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAStock(t *testing.T) {
	shouldBeTrue := []string{"AAA", "A32 ", "VN30F2010'"}
	for _, v := range shouldBeTrue {
		cleaned, isStock := isAStock(v)
		assert.Equal(t, true, isStock, cleaned+" should be a stock")
	}
	shouldBeFalse := []string{"000", "AAAAAAAAAAAAAAAAAAAAAA", "2AB", "�"}
	for _, v := range shouldBeFalse {
		cleaned, isStock := isAStock(v)
		assert.Equal(t, false, isStock, cleaned+" should not be a stock")
	}
}
