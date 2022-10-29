package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBogusCoinReplacement(t *testing.T) {
	for _, tc := range []struct {
		desc string
		msg  string
		exp  string
	}{
		{
			desc: "Single match",
			msg:  "Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX , and yes\n",
			exp:  "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI , and yes\n",
		},
		{
			desc: "Multi match",
			msg:  "Hi alice this is my 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX , and please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
			exp:  "Hi alice this is my 7YWHMfk9JZe0LM0g1ZauHuiSxhI , and please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n",
		},
		{
			desc: "should not match first one due to comma",
			msg:  "Hi alice this is my7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX, and please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
			exp:  "Hi alice this is my7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX, and please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n",
		},
		{
			desc: "should not match 2nd one due to front comma",
			msg:  "Hi alice this is 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX , and please send payment ,to7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
			exp:  "Hi alice this is 7YWHMfk9JZe0LM0g1ZauHuiSxhI , and please send payment ,to7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
		},
		{
			desc: "should not match any due to commas",
			msg:  "Hi alice this is 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX, and please send payment ,to7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
			exp:  "Hi alice this is 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX, and please send payment ,to7iKDZEwPZSqIvDnHvVN2r0hUD5rHX\n",
		},
		{
			desc: "should not match",
			msg:  "Hi alice\n",
			exp:  "Hi alice\n",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			rewritten := rewrite([]byte(tc.msg))
			require.Equal(t, tc.exp, string(rewritten))
		})
	}

}
