// Package ml provides small, dependency-free machine-learning primitives used by
// the analytics service to turn warehouse aggregates into predictions and
// segments. Everything here is pure Go (no external libraries) and fully
// deterministic so the same warehouse snapshot always yields the same output.
package ml

import (
	"math"
	"sort"
)

// ---------------------------------------------------------------------------
// Ordinary least-squares linear regression:  y = Slope*x + Intercept
// Used for demand/revenue forecasting and residual-based anomaly detection.
// ---------------------------------------------------------------------------

type Line struct {
	Slope     float64 `json:"slope"`
	Intercept float64 `json:"intercept"`
	R2        float64 `json:"r2"`        // goodness of fit, 0..1
	StdResid  float64 `json:"std_resid"` // population std-dev of residuals
}

// FitLine fits a least-squares line to y over evenly spaced x = 0..len(y)-1.
func FitLine(y []float64) Line {
	n := len(y)
	if n == 0 {
		return Line{}
	}
	if n == 1 {
		return Line{Intercept: y[0], R2: 1}
	}
	nf := float64(n)
	var sx, sy, sxx, sxy float64
	for i, v := range y {
		x := float64(i)
		sx += x
		sy += v
		sxx += x * x
		sxy += x * v
	}
	denom := nf*sxx - sx*sx
	if denom == 0 {
		return Line{Intercept: sy / nf}
	}
	slope := (nf*sxy - sx*sy) / denom
	intercept := (sy - slope*sx) / nf

	mean := sy / nf
	var ssTot, ssRes float64
	for i, v := range y {
		r := v - (slope*float64(i) + intercept)
		ssRes += r * r
		ssTot += (v - mean) * (v - mean)
	}
	r2 := 1.0
	if ssTot > 0 {
		r2 = 1 - ssRes/ssTot
	}
	return Line{Slope: slope, Intercept: intercept, R2: r2, StdResid: math.Sqrt(ssRes / nf)}
}

// At evaluates the fitted line at position x.
func (l Line) At(x float64) float64 { return l.Slope*x + l.Intercept }

// ---------------------------------------------------------------------------
// K-Means clustering (Lloyd's algorithm) with deterministic seeding.
// Used to segment destination provinces into tiers from multi-feature data.
// ---------------------------------------------------------------------------

// standardize z-scores each column so features on different scales (package
// counts vs. rupiah) contribute equally to the Euclidean distance.
func standardize(rows [][]float64) [][]float64 {
	if len(rows) == 0 {
		return rows
	}
	d := len(rows[0])
	n := float64(len(rows))
	means := make([]float64, d)
	stds := make([]float64, d)
	for _, r := range rows {
		for j := 0; j < d; j++ {
			means[j] += r[j]
		}
	}
	for j := range means {
		means[j] /= n
	}
	for _, r := range rows {
		for j := 0; j < d; j++ {
			dv := r[j] - means[j]
			stds[j] += dv * dv
		}
	}
	for j := range stds {
		stds[j] = math.Sqrt(stds[j] / n)
		if stds[j] == 0 {
			stds[j] = 1
		}
	}
	out := make([][]float64, len(rows))
	for i, r := range rows {
		nr := make([]float64, d)
		for j := 0; j < d; j++ {
			nr[j] = (r[j] - means[j]) / stds[j]
		}
		out[i] = nr
	}
	return out
}

func dist2(a, b []float64) float64 {
	var s float64
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return s
}

// KMeans groups rows into k clusters and returns the cluster index per row.
// Seeding is deterministic: features are standardized, points ranked by their
// summed score, and k evenly-spaced points become the initial centroids. This
// avoids any RNG so results are reproducible across runs.
func KMeans(rows [][]float64, k, iters int) []int {
	n := len(rows)
	assign := make([]int, n)
	if n == 0 {
		return assign
	}
	if k > n {
		k = n
	}
	if k <= 1 {
		return assign
	}

	z := standardize(rows)
	dcol := len(z[0])

	order := make([]int, n)
	score := make([]float64, n)
	for i := range order {
		order[i] = i
		for _, v := range z[i] {
			score[i] += v
		}
	}
	sort.Slice(order, func(a, b int) bool { return score[order[a]] < score[order[b]] })

	centroids := make([][]float64, k)
	for c := 0; c < k; c++ {
		idx := order[c*(n-1)/(k-1)] // spread across the sorted range
		centroids[c] = append([]float64(nil), z[idx]...)
	}

	for it := 0; it < iters; it++ {
		changed := false
		for i := range z {
			best, bd := 0, math.MaxFloat64
			for c := range centroids {
				if d := dist2(z[i], centroids[c]); d < bd {
					bd, best = d, c
				}
			}
			if assign[i] != best {
				assign[i], changed = best, true
			}
		}
		sums := make([][]float64, k)
		counts := make([]int, k)
		for c := range sums {
			sums[c] = make([]float64, dcol)
		}
		for i, r := range z {
			c := assign[i]
			counts[c]++
			for j := 0; j < dcol; j++ {
				sums[c][j] += r[j]
			}
		}
		for c := range centroids {
			if counts[c] == 0 {
				continue
			}
			for j := 0; j < dcol; j++ {
				centroids[c][j] = sums[c][j] / float64(counts[c])
			}
		}
		if !changed && it > 0 {
			break
		}
	}
	return assign
}
