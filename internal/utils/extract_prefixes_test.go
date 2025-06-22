package utils

import "testing"

func TestExtractPrefixes(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name            string
		args            args
		wantDepth3      string
		wantDepth2      string
		wantDepth1      string
		isMoreThanThree bool
	}{
		{
			name:       "root path",
			args:       args{path: "/"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "", isMoreThanThree: false,
		},
		{
			name:       "root path",
			args:       args{path: "r"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "/r", isMoreThanThree: false,
		},
		{
			name:       "single segment with slash",
			args:       args{path: "/api"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "two segments no trailing slash",
			args:       args{path: "/api/v1"},
			wantDepth3: "", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "two segments with trailing slash",
			args:       args{path: "/api/v1/"},
			wantDepth3: "", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "three segments no trailing slash",
			args:       args{path: "/api/v1/users"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "three segments with trailing slash",
			args:       args{path: "/api/v1/users/"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "three segments with query string",
			args:       args{path: "/api/v1/users?id=10"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "missing leading slash",
			args:       args{path: "api/v1/users"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "only one segment",
			args:       args{path: "/onlyone/"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "/onlyone", isMoreThanThree: false,
		},
		{
			name:       "auth single",
			args:       args{path: "/auth"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "/auth", isMoreThanThree: false,
		},
		{
			name:       "auth versioned",
			args:       args{path: "/auth/v2"},
			wantDepth3: "", wantDepth2: "/auth/v2", wantDepth1: "/auth", isMoreThanThree: false,
		},
		{
			name:       "auth versioned trailing slash",
			args:       args{path: "/auth/v2/"},
			wantDepth3: "", wantDepth2: "/auth/v2", wantDepth1: "/auth", isMoreThanThree: false,
		},
		{
			name:       "empty string",
			args:       args{path: ""},
			wantDepth3: "", wantDepth2: "", wantDepth1: "", isMoreThanThree: false,
		},
		// multiple slashes not supported
		//{
		//	name:       "double slashes",
		//	args:       args{path: "/api//v1"},
		//	wantDepth3: "", wantDepth2: "/api", wantDepth1: "/api", isMoreThanThree: false,
		//},
		//{
		//	name:       "multiple slashes",
		//	args:       args{path: "///api///v1///users"},
		//	wantDepth3: "/api", wantDepth2: "/api", wantDepth1: "/api", isMoreThanThree: false,
		//},
		{
			name:       "trailing slash with query",
			args:       args{path: "/api/v1/users/?sort=asc"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: false,
		},
		{
			name:       "query with no path",
			args:       args{path: "?id=5"},
			wantDepth3: "", wantDepth2: "", wantDepth1: "", isMoreThanThree: false,
		},
		{
			name:       "deep path beyond 3 levels",
			args:       args{path: "/a/b/c/d/e"},
			wantDepth3: "/a/b/c", wantDepth2: "/a/b", wantDepth1: "/a", isMoreThanThree: true,
		},
		{
			name:       "deep path with no trailing slash",
			args:       args{path: "/api/v1/users/extra/more"},
			wantDepth3: "/api/v1/users", wantDepth2: "/api/v1", wantDepth1: "/api", isMoreThanThree: true,
		},
		{
			name:       "single letter segments",
			args:       args{path: "/a/b/c"},
			wantDepth3: "/a/b/c", wantDepth2: "/a/b", wantDepth1: "/a", isMoreThanThree: false,
		},
		{
			name:       "long segment names",
			args:       args{path: "/this-is-a/very-long/path-name"},
			wantDepth3: "/this-is-a/very-long/path-name", wantDepth2: "/this-is-a/very-long", wantDepth1: "/this-is-a", isMoreThanThree: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDepth3, gotDepth2, gotDepth1, isMoreThanThree := ExtractPrefixes(tt.args.path)
			if gotDepth3 != tt.wantDepth3 {
				t.Errorf("ExtractPrefixes() gotDepth3 = %v, want %v", gotDepth3, tt.wantDepth3)
			}
			if gotDepth2 != tt.wantDepth2 {
				t.Errorf("ExtractPrefixes() gotDepth2 = %v, want %v", gotDepth2, tt.wantDepth2)
			}
			if gotDepth1 != tt.wantDepth1 {
				t.Errorf("ExtractPrefixes() gotDepth1 = %v, want %v", gotDepth1, tt.wantDepth1)
			}
			if isMoreThanThree != tt.isMoreThanThree {
				t.Errorf("ExtractPrefixes() isMoreThanThree = %v, want %v", isMoreThanThree, tt.isMoreThanThree)
			}
		})
	}
}
