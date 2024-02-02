/*
Copyright 2024 The Kubernetes Authors.

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

package strings

import (
	"testing"
)

func TestCompareLists(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "Empty lists",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "Equal lists",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "Different order",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "b", "a"},
			want: true,
		},
		{
			name: "Different elements",
			a:    []string{"a", "b", "c"},
			b:    []string{"d", "e", "f"},
			want: false,
		},
		{
			name: "Different lengths",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareLists(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareLists() = %v, want %v", got, tt.want)
			}
		})
	}
}
