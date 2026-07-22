package reset

import (
	"testing"
)

// Тестовый тип, который имплементирует интерфейс Resetable
type TestStruct struct {
	Value int
}

func (ts *TestStruct) Reset() {
	ts.Value = 0
}

func TestNew(t *testing.T) {
	pool := New[*TestStruct]()
	if pool == nil {
		t.Errorf("Expected non-nil Pool, got nil")
	}
}

func TestGetAndPut(t *testing.T) {
	pool := New[*TestStruct]()

	// Проверяем Get() при пустом пуле
	obj := pool.Get()
	if obj == nil {
		obj = &TestStruct{}
	}
	if obj.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", obj.Value)
	}

	// Помещаем объект в пул и проверяем его
	pool.Put(obj)
	newObj := pool.Get()
	if newObj.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", newObj.Value)
	}
}

func TestPutResetsObject(t *testing.T) {
	pool := New[*TestStruct]()
	testObj := &TestStruct{Value: 42}
	pool.Put(testObj)

	newObj := pool.Get()
	if newObj.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", newObj.Value)
	}
}

func TestPoolGrows(t *testing.T) {
	pool := New[*TestStruct]()
	testObj1 := &TestStruct{Value: 42}
	testObj2 := &TestStruct{Value: 84}

	pool.Put(testObj1)
	pool.Put(testObj2)

	newObj1 := pool.Get()
	if newObj1.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", newObj1.Value)
	}

	newObj2 := pool.Get()
	if newObj2.Value != 0 {
		t.Errorf("Expected Value to be 0, got %d", newObj2.Value)
	}
}
