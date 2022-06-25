package main

import (
	"bit-wordy/src/cached"
	"bit-wordy/src/primitives"
	"testing"
)

func TestIterate_Run(t *testing.T) {
	type fields struct {
		Times int
	}
	type args struct {
		p *cached.Patterns
	}
	a := args{}
	var err error
	a.p, err = cached.LoadPatterns(primitives.LoadWords())
	if err != nil {
		t.Errorf("Could not load patterns")
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"Answer: girts", fields{10}, a, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := Iterate{
				Times: tt.fields.Times,
			}
			if err := i.Run(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
