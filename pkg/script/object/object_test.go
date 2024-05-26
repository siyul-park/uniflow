package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys: %#v != %#v",
			hello1.HashKey(), hello2.HashKey())
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys: %#v != %#v",
			diff1.HashKey(), diff2.HashKey())
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys: %#v != %#v",
			hello1.HashKey(), diff1.HashKey())
	}
}

func TestBooleanHashKey(t *testing.T) {
	true1 := &Boolean{Value: true}
	true2 := &Boolean{Value: true}
	false1 := &Boolean{Value: false}
	false2 := &Boolean{Value: false}

	if true1.HashKey() != true2.HashKey() {
		t.Errorf("trues do not have the same hash keys: %#v != %#v",
			true1.HashKey(), true2.HashKey())
	}

	if false1.HashKey() != false2.HashKey() {
		t.Errorf("falses do not have the same hash keys: %#v != %#v",
			false1.HashKey(), false2.HashKey())
	}

	if true1.HashKey() == false1.HashKey() {
		t.Errorf("true has same hash key as false: %#v != %#v",
			true1.HashKey(), false1.HashKey())
	}
}

func TestIntegerHashKey(t *testing.T) {
	one1 := &Integer{Value: 1}
	one2 := &Integer{Value: 1}
	two1 := &Integer{Value: 2}
	two2 := &Integer{Value: 2}

	if one1.HashKey() != one2.HashKey() {
		t.Errorf("integers with same content have different hash keys: %#v != %#v",
			one1.HashKey(), one2.HashKey())
	}

	if two1.HashKey() != two2.HashKey() {
		t.Errorf("integers with same content have different hash keys: %#v != %#v",
			two1.HashKey(), two2.HashKey())
	}

	if one1.HashKey() == two1.HashKey() {
		t.Errorf("integers with different content have same hash keys: %#v != %#v",
			one1.HashKey(), two1.HashKey())
	}
}

func TestNilHashKey(t *testing.T) {
	n1 := &Nil{}
	n2 := &Nil{}

	if n1.HashKey() != n2.HashKey() {
		t.Errorf("nils have different hash keys: %#v != %#v", n1.HashKey(), n2.HashKey())
	}
}
