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

type LayeredMap struct {
	Layers        []map[string]string
	WhiteoutFiles map[string]bool
	hasher        func(string) (string, error)
}

func NewLayeredMap(h func(string) (string, error)) *LayeredMap {
	l := LayeredMap{
		hasher: h,
	}
	l.Layers = []map[string]string{}
	l.WhiteoutFiles = make(map[string]bool)
	return &l
}

func (l *LayeredMap) Snapshot() {
	l.Layers = append(l.Layers, map[string]string{})
}

func (l *LayeredMap) Get(s string) (string, bool) {
	for i := len(l.Layers) - 1; i >= 0; i-- {
		if v, ok := l.Layers[i][s]; ok {
			return v, ok
		}
	}
	return "", false
}

func (l *LayeredMap) MaybeAdd(s string) (bool, error) {
	oldV, ok := l.Get(s)
	newV, err := l.hasher(s)
	if err != nil {
		return false, err
	}
	if ok && newV == oldV {
		return false, nil
	}
	l.Layers[len(l.Layers)-1][s] = newV
	return true, nil
}
