package color15

import "testing"

func TestMulDiv(t *testing.T) {
	type args struct {
		r            uint8
		g            uint8
		b            uint8
		multiplicand uint8
		divisor      uint8
	}
	tests := []struct {
		name   string
		args   args
		wantMr uint8
		wantMg uint8
		wantMb uint8
	}{
		{
			name: "12ef",
			args: args{
				r:            15,
				g:            23,
				b:            4,
				multiplicand: 25,
				divisor:      31,
			},
			wantMr: 12,
			wantMg: 18,
			wantMb: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToColor15(tt.args.r, tt.args.g, tt.args.b).MulDiv(tt.args.multiplicand, tt.args.divisor)
			r, g, b := got.ToRGB()
			if r != tt.wantMr {
				t.Errorf("MulDiv() gotMr = %v, want %v", r, tt.wantMr)
			}
			if g != tt.wantMg {
				t.Errorf("MulDiv() gotMg = %v, want %v", g, tt.wantMg)
			}
			if b != tt.wantMb {
				t.Errorf("MulDiv() gotMb = %v, want %v", b, tt.wantMb)
			}
		})
	}
}
