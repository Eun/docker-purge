package jq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchFilter(t *testing.T) {
	tests := []struct {
		Input    string
		Filter   string
		Ok       bool
		HasError bool
	}{
		{
			`{"Name": "Joe", "IsMale": true}`,
			`.IsMale==true`,
			true,
			false,
		},
		{
			`{"Name": "Joe", "IsMale": true}`,
			`.IsMale==false`,
			false,
			false,
		},
		{
			`{"Name": "Joe", "IsMale": true}`,
			`.IsMale==true and (.Name | contains("J"))`,
			true,
			false,
		},
	}
	for _, test := range tests {
		ok, err := MatchesFilter(test.Input, test.Filter)
		require.Equal(t, test.Ok, ok)
		if test.HasError {
			require.NotNil(t, err, "Expected Error")
		} else {
			require.Nil(t, err, "Expected no Error")
		}
	}
}
