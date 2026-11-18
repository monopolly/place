package model

import (
	"fmt"
	"strings"
)

func (a *P) ReindexOptions() {

	var options []string
	options = append(options, fmt.Sprintf("@%s", a.Country()))
	if a.State() != "" {
		options = append(options, fmt.Sprintf("@%s", a.State()))
	}

	for _, x := range strings.Fields(a.Tags()) {
		if !strings.HasPrefix(x, "@") {
			x = "@" + x
		}
		options = append(options, x)
	}

	if a.Zip() != "" {
		options = append(options, fmt.Sprintf("%s", a.Zip()))
	}

	a.O(options...)

}
