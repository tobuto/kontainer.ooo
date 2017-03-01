package testutils

import (
	"errors"
	"fmt"
	"reflect"
)

// Table simulates a Table in the MockDb
type table struct {
	Name  string
	rows  []interface{}
	ref   reflect.Type
	idx   bool
	count uint
}

func (t *table) String() string {
	s := fmt.Sprintf("Table: %s\n", t.Name)
	for i, r := range t.rows {
		s += fmt.Sprintf("ID: %d - %v\n", i, r)
	}
	return s
}

func (t *table) copy() *table {
	nt := &table{
		Name:  t.Name,
		ref:   t.ref,
		idx:   t.idx,
		count: t.count,
	}

	for _, r := range t.rows {
		nt.rows = append(nt.rows, *(&r))
	}

	return nt
}

func (t *table) checkForField(f string) bool {
	_, found := t.ref.FieldByName(f)
	return found
}

func (t *table) typeCheck(i interface{}) error {
	typ := reflect.TypeOf(i)
	if typ != reflect.PtrTo(t.ref) && typ != t.ref {
		return ErrTypeMismatch
	}
	return nil
}

func (t *table) insert(row interface{}) error {
	err := t.typeCheck(row)
	if err != nil {
		return err
	}

	t.count++
	if t.idx {
		s := reflect.ValueOf(row).Elem()
		idField := s.FieldByName("ID")
		value := reflect.ValueOf(t.count)
		idField.Set(value)
	}
	t.rows = append(t.rows, row)
	return nil
}

func (t *table) find(field string, value interface{}) (reflect.Value, error) {
	_, found := t.ref.FieldByName(field)
	if !found {
		return reflect.ValueOf(nil), errors.New("field name not in struct")
	}
	result := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(t.ref)), 0, 0)
	for _, row := range t.rows {
		if reflect.TypeOf(value).Kind() == reflect.Uint {
			v := reflect.ValueOf(row).Elem().FieldByName(field).Uint()
			if uint(v) == value {
				result = reflect.Append(result, reflect.ValueOf(row))
			}
		} else {
			v := reflect.ValueOf(row).Elem().FieldByName(field).String()
			if v == value {
				result = reflect.Append(result, reflect.ValueOf(row))
			}
		}
	}

	if result.Len() == 0 {
		return reflect.ValueOf(nil), nil
	}
	return result, nil
}

func (t *table) delete(id uint64) error {
	for i, row := range t.rows {
		v := reflect.ValueOf(row).Elem().FieldByName("ID").Uint()
		if v == id {
			t.rows = append(t.rows[:i], t.rows[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func newTable(ref reflect.Type, name string) *table {
	_, found := ref.FieldByName("ID")
	return &table{
		Name: name,
		ref:  ref,
		idx:  found,
	}
}