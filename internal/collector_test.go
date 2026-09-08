package internal

import (
	"reflect"
	"slices"
	"testing"
	"time"
)

func Test_generateDateRange(t *testing.T) {
	type args struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		name string
		args args
		want []time.Time
	}{
		{
			"success 1 day",
			args{
				from: time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC),
				to:   time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC),
			},
			[]time.Time{time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC)},
		},
		{
			"success 4 days",
			args{
				from: time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC),
				to:   time.Date(2020, time.March, 13, 0, 0, 0, 0, time.UTC),
			},
			[]time.Time{
				time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.March, 11, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.March, 12, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.March, 13, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"from > to",
			args{
				from: time.Date(2020, time.March, 13, 0, 0, 0, 0, time.UTC),
				to:   time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC),
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDateRange(tt.args.from, tt.args.to)
			gotSlice := slices.Collect(got)
			if !reflect.DeepEqual(gotSlice, tt.want) {
				t.Errorf("generateDateRange() = %v, want %v", gotSlice, tt.want)
			}
		})
	}
}
