package money

import "testing"

func TestFromFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected Money
	}{
		{10.50, 1050},
		{100.00, 10000},
		{0.01, 1},
		{99.99, 9999},
	}

	for _, tt := range tests {
		result := FromFloat(tt.input)
		if result != tt.expected {
			t.Errorf("FromFloat(%f) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		input    Money
		expected float64
	}{
		{1050, 10.50},
		{10000, 100.00},
		{1, 0.01},
		{9999, 99.99},
	}

	for _, tt := range tests {
		result := tt.input.Float()
		if result != tt.expected {
			t.Errorf("Money(%d).Float() = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    Money
		expected string
	}{
		{FromFloat(10.50), "10.50"},
		{FromFloat(100.00), "100.00"},
		{FromFloat(0.01), "0.01"},
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("Money.String() = %s, want %s", result, tt.expected)
		}
	}
}

func TestAdd(t *testing.T) {
	a := FromFloat(10.50)
	b := FromFloat(5.25)
	result := a.Add(b)
	expected := FromFloat(15.75)

	if result != expected {
		t.Errorf("Add() = %s, want %s", result.String(), expected.String())
	}
}

func TestSub(t *testing.T) {
	a := FromFloat(10.50)
	b := FromFloat(5.25)
	result := a.Sub(b)
	expected := FromFloat(5.25)

	if result != expected {
		t.Errorf("Sub() = %s, want %s", result.String(), expected.String())
	}
}

func TestMul(t *testing.T) {
	m := FromFloat(10.00)
	result := m.Mul(1.5)
	expected := FromFloat(15.00)

	if result != expected {
		t.Errorf("Mul(1.5) = %s, want %s", result.String(), expected.String())
	}
}

func TestAlmostEqual(t *testing.T) {
	a := FromFloat(10.00)
	b := FromFloat(10.01)
	tolerance := FromFloat(0.02)

	if !AlmostEqual(a, b, tolerance) {
		t.Error("AlmostEqual() should return true")
	}

	tolerance = FromFloat(0.001)
	if AlmostEqual(a, b, tolerance) {
		t.Error("AlmostEqual() should return false")
	}
}
