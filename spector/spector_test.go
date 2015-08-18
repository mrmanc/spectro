package main
import (
	"testing"
)
func TestExponentialScale(t *testing.T) {
	assertEqualFloat(t, exponentialScale(0, 1234, 10), 0)
	assertEqualFloat(t, exponentialScale(10, 1024, 10), 1024)
	assertEqualFloat(t, exponentialScale(10, 1234, 10), 1234)
}
func TestLogarithmicScale(t *testing.T) {
	assertEqualFloat(t, logarithmicScale(0, 1234, 10), 0)
	assertEqualFloat(t, logarithmicScale(10, 1024, 10), 1024)
	assertEqualFloat(t, logarithmicScale(10, 1234, 10), 1234)
}
func TestExponentialReversability(t *testing.T) {
	assertEqualInt(t, reverseExponentialScale(exponentialScale(0, 1024, 10), 1024, 10), 0)
	assertEqualInt(t, reverseExponentialScale(exponentialScale(2, 1024, 10), 1024, 10), 2)
	assertEqualInt(t, reverseExponentialScale(exponentialScale(5, 1024, 10), 1024, 10), 5)
	assertEqualInt(t, reverseExponentialScale(exponentialScale(7, 1024, 10), 1024, 10), 7)
	assertEqualInt(t, reverseExponentialScale(exponentialScale(10, 1024, 10), 1024, 10), 10)
}
func TestLogarithmicReversability(t *testing.T) {
	assertEqualInt(t, reverseLogarithmicScale(logarithmicScale(0, 1024, 10), 1024, 10), 0)
	assertEqualInt(t, reverseLogarithmicScale(logarithmicScale(2, 1024, 10), 1024, 10), 2)
	assertEqualInt(t, reverseLogarithmicScale(logarithmicScale(5, 1024, 10), 1024, 10), 5)
	assertEqualInt(t, reverseLogarithmicScale(logarithmicScale(7, 1024, 10), 1024, 10), 7)
	assertEqualInt(t, reverseLogarithmicScale(logarithmicScale(10, 1024, 10), 1024, 10), 10)
}
func TestLinearReversability(t *testing.T) {
	assertEqualInt(t, reverseLinearScale(linearScale(0, 1024, 10), 1024, 10), 0)
	assertEqualInt(t, reverseLinearScale(linearScale(2, 1024, 10), 1024, 10), 2)
	assertEqualInt(t, reverseLinearScale(linearScale(5, 1024, 10), 1024, 10), 5)
	assertEqualInt(t, reverseLinearScale(linearScale(7, 1024, 10), 1024, 10), 7)
	assertEqualInt(t, reverseLinearScale(linearScale(10, 1024, 10), 1024, 10), 10)
	assertEqualInt(t, reverseLinearScale(linearScale(5, 321, 13), 321, 13), 5)
}
func assertEqualInt(t *testing.T, got uint, expected uint) {
	if (got != expected) {
		t.Errorf("got %v, wanted %v", got, expected)
	}
}
func assertEqualFloat(t *testing.T, got float64, expected float64) {
	if (!floatEquals(got,expected)) {
		t.Errorf("got %v, wanted %v", got, expected)
	}
}
var EPSILON float64 = 0.00000001
func floatEquals(a, b float64) bool {
	if ((a - b) < EPSILON && (b - a) < EPSILON) {
		return true
	}
	return false
}