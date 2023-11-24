package main

import "testing"

func TestVisitSite(t *testing.T) {
	html := visitSite("https://22bet.ng/en/live/football", 40)
	dom := createDOM(html)
	matchEvents := SeperateObjects(dom)
	for _, v := range matchEvents {
		t.Error(v)
	}

}
