/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snapshot

import (
	"fmt"
	"path/filepath"
	"strings"
)

type LayeredMap struct {
	layers    []map[string]string
	whiteouts []map[string]string
	hasher    func(string) (string, error)
}

func NewLayeredMap(h func(string) (string, error)) *LayeredMap {
	l := LayeredMap{
		hasher: h,
	}
	l.layers = []map[string]string{}
	return &l
}

func (l *LayeredMap) Snapshot() {
	l.whiteouts = append(l.whiteouts, map[string]string{})
	l.layers = append(l.layers, map[string]string{})
}

func (l *LayeredMap) GetFlattenedPathsForWhiteOut() map[string]struct{} {
	paths := map[string]struct{}{}
	for _, l := range l.layers {
		for p := range l {
			if strings.HasPrefix(filepath.Base(p), ".wh.") {
				delete(paths, p)
			} else {
				paths[p] = struct{}{}
			}
			paths[p] = struct{}{}
		}
	}
	return paths
}

func (l *LayeredMap) Get(s string) (string, bool) {
	for i := len(l.layers) - 1; i >= 0; i-- {
		if v, ok := l.layers[i][s]; ok {
			return v, ok
		}
	}
	return "", false
}

func (l *LayeredMap) GetWhiteout(s string) (string, bool) {
	for i := len(l.whiteouts) - 1; i >= 0; i-- {
		if v, ok := l.whiteouts[i][s]; ok {
			return v, ok
		}
	}
	return "", false
}

func (l *LayeredMap) MaybeAddWhiteout(s string) (bool, error) {
	whiteout, ok := l.GetWhiteout(s)
	if ok && whiteout == s {
		return false, nil
	}
	l.whiteouts[len(l.whiteouts)-1][s] = s
	return true, nil
}

// Add will add the specified file s to the layered map.
func (l *LayeredMap) Add(s string) error {
	newV, err := l.hasher(s)
	if err != nil {
		return fmt.Errorf("Error creating hash for %s: %s", s, err)
	}
	l.layers[len(l.layers)-1][s] = newV
	return nil
}

// MaybeAdd will add the specified file s to the layered map if
// the layered map's hashing function determines it has changed. If
// it has not changed, it will not be added. Returns true if the file
// was added.
func (l *LayeredMap) MaybeAdd(s string) (bool, error) {
	oldV, ok := l.Get(s)
	newV, err := l.hasher(s)
	if err != nil {
		return false, err
	}
	if ok && newV == oldV {
		return false, nil
	}
	l.layers[len(l.layers)-1][s] = newV
	return true, nil
}
