package model_test

import (
	"testing"

	"github.com/yandex-practicum/shorten-url/internal/model"
)

func TestResetableStruct_Reset(t *testing.T) {
	strVal := "hello"
	initialChild := &model.ResetableStruct{I: 100, STR: "child"}

	rs := &model.ResetableStruct{
		I:     42,
		STR:   "main",
		STRP:  &strVal,
		S:     []int{1, 2, 3, 4, 5},
		M:     map[string]string{"key": "value"},
		CHILD: initialChild,
	}

	oldCap := cap(rs.S)

	rs.Reset()

	// Примитивы
	if rs.I != 0 {
		t.Errorf("i: ожидалось 0, получено %d", rs.I)
	}
	if rs.STR != "" {
		t.Errorf("str: ожидалось \"\", получено %q", rs.STR)
	}

	// Указатель на примитив
	if rs.STRP != nil && *rs.STRP != "" {
		t.Errorf("strP: значение по адресу должно быть \"\", получено %q", *rs.STRP)
	}

	// Слайс
	if len(rs.S) != 0 {
		t.Errorf("s: длина должна быть 0, получено %d", len(rs.S))
	}
	if cap(rs.S) != oldCap {
		t.Errorf("s: емкость должна сохраниться %d, получено %d", oldCap, cap(rs.S))
	}

	// Мапа
	if len(rs.M) != 0 {
		t.Errorf("m: мапа должна быть пустой, получено %d элементов", len(rs.M))
	}

	// Вложенная структура
	if rs.CHILD != nil {
		if rs.CHILD.I != 0 || rs.CHILD.STR != "" {
			t.Error("child: метод Reset() не был вызван рекурсивно для вложенной структуры")
		}
	}
}

func TestReset_NilReceiver(t *testing.T) {
	var rs *model.ResetableStruct
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Reset() запаниковал на nil ресивере: %v", r)
		}
	}()
	rs.Reset()
}
