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
	"testing"
)

func Test_Key(t *testing.T) {

	map1 := map[string]string{
		"hey": "hi",
		"hi":  "hey",
	}
	map2 := map[string]string{
		"hi":  "hey",
		"hey": "hi",
	}

	lm1 := LayeredMap{
		layers: []map[string]string{map1},
	}

	lm2 := LayeredMap{
		layers: []map[string]string{map2},
	}
	key1, err := lm1.Key()
	if err != nil {
		t.Fatal(err)
	}
	key2, err := lm2.Key()
	if err != nil {
		t.Fatal(err)
	}
	if key1 == key2 {
		t.Fatalf("sad, key1: %s, kedy2: %s", key1, key2)
	}
}
