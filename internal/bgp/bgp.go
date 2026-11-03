// Package bgp demonstrates simplified BGP best-path selection on sample route data.
// This is NOT a BGP daemon, parser, peering implementation, or router configuration tool.
package bgp

import (
	"errors"
	"net/netip"
	"sort"
)

type Route struct {
	Prefix    string   `json:"prefix"`
	NextHop   string   `json:"next_hop"`
	LocalPref int      `json:"local_pref"`
	ASPath    []uint32 `json:"as_path"`
	MED       int      `json:"med"`
	Origin    string   `json:"origin"`
}

func valid(r Route) error {
	if _, err := netip.ParsePrefix(r.Prefix); err != nil {
		return err
	}
	if _, err := netip.ParseAddr(r.NextHop); err != nil {
		return err
	}
	if r.LocalPref < 0 || r.MED < 0 {
		return errors.New("negative routing attribute")
	}
	return nil
}

// Best selects deterministically: highest local preference, shortest AS path,
// lowest MED (lab simplification: compare across all peers), then next-hop lexical order.
// Real BGP has additional attributes and rules, including restrictions on MED comparisons.
func Best(routes []Route) (Route, error) {
	if len(routes) == 0 {
		return Route{}, errors.New("no routes")
	}
	prefix := routes[0].Prefix
	copyRoutes := append([]Route(nil), routes...)
	for _, r := range copyRoutes {
		if err := valid(r); err != nil {
			return Route{}, err
		}
		if r.Prefix != prefix {
			return Route{}, errors.New("routes must share prefix")
		}
	}
	sort.Slice(copyRoutes, func(i, j int) bool {
		a, b := copyRoutes[i], copyRoutes[j]
		if a.LocalPref != b.LocalPref {
			return a.LocalPref > b.LocalPref
		}
		if len(a.ASPath) != len(b.ASPath) {
			return len(a.ASPath) < len(b.ASPath)
		}
		if a.MED != b.MED {
			return a.MED < b.MED
		}
		return a.NextHop < b.NextHop
	})
	return copyRoutes[0], nil
}
