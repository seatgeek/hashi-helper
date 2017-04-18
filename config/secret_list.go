package config

import "sort"

// SecretList ...
type SecretList []*InternalSecret

func (p SecretList) Len() int { return len(p) }
func (p SecretList) Less(i, j int) bool {
	return p[i].Environment+"_"+p[i].Application+"_"+p[i].Path < p[j].Environment+"_"+p[j].Application+"_"+p[j].Path
}
func (p SecretList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p SecretList) Sort() { sort.Sort(p) }
