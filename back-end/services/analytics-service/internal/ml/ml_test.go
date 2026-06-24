package ml

import (
	"math"
	"testing"
)

func TestFitLinePerfectTrend(t *testing.T) {
	// y = 2x + 1 → slope 2, intercept 1, R²=1
	l := FitLine([]float64{1, 3, 5, 7, 9})
	if math.Abs(l.Slope-2) > 1e-9 || math.Abs(l.Intercept-1) > 1e-9 {
		t.Fatalf("slope/intercept = %v/%v, want 2/1", l.Slope, l.Intercept)
	}
	if math.Abs(l.R2-1) > 1e-9 {
		t.Fatalf("R2 = %v, want 1", l.R2)
	}
	if got := l.At(5); math.Abs(got-11) > 1e-9 {
		t.Fatalf("At(5) = %v, want 11", got)
	}
}

func TestKMeansSeparatesClusters(t *testing.T) {
	// Two tight, well-separated blobs → k=2 must split them.
	rows := [][]float64{
		{0, 0}, {1, 0}, {0, 1}, {1, 1},
		{50, 50}, {51, 50}, {50, 51}, {51, 51},
	}
	a := KMeans(rows, 2, 50)
	if a[0] == a[4] {
		t.Fatalf("expected the two blobs in different clusters, got %v", a)
	}
	for i := 1; i < 4; i++ {
		if a[i] != a[0] {
			t.Fatalf("first blob not cohesive: %v", a)
		}
	}
	for i := 5; i < 8; i++ {
		if a[i] != a[4] {
			t.Fatalf("second blob not cohesive: %v", a)
		}
	}
}

func TestKMeansDeterministic(t *testing.T) {
	rows := [][]float64{{1, 2}, {2, 1}, {9, 9}, {8, 8}, {1, 1}, {10, 9}}
	if a, b := KMeans(rows, 3, 50), KMeans(rows, 3, 50); !equal(a, b) {
		t.Fatalf("k-means not deterministic: %v vs %v", a, b)
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
