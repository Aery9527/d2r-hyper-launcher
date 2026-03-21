package d2r

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindRegion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NA", "us.actual.battle.net"},
		{"na", "us.actual.battle.net"},
		{"EU", "eu.actual.battle.net"},
		{"eu", "eu.actual.battle.net"},
		{"Asia", "kr.actual.battle.net"},
		{"asia", "kr.actual.battle.net"},
	}

	for _, tt := range tests {
		r := FindRegion(tt.input)
		assert.NotNil(t, r, "region %q should be found", tt.input)
		assert.Equal(t, tt.expected, r.Address)
	}
}

func TestFindRegion_NotFound(t *testing.T) {
	r := FindRegion("XX")
	assert.Nil(t, r)
}

func TestNormalizeRegionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NA", "NA"},
		{"na", "NA"},
		{" EU ", "EU"},
		{"asia", "Asia"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, NormalizeRegionName(tt.input))
	}
}

func TestNormalizeRegionName_NotFound(t *testing.T) {
	assert.Equal(t, "", NormalizeRegionName("XX"))
}
