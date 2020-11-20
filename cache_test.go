package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func genValue(num int) Value {
	return fmt.Sprintf("value%v", num)
}

func TestInMemoryCache_GetOrSet(t *testing.T) {
	var gen uint8

	type fields struct {
		dataMutex sync.RWMutex
		data      map[Key]Value
	}
	type args struct {
		key     Key
		valueFn func() Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Value
	}{
		{
			name:   "Проверяем что вернет GetOrSet",
			fields: fields{
				dataMutex: sync.RWMutex{},
				data: map[Key]Value{
					"key1": "value1",
				},
			},
			args:   args{
				key:     "k1",
				valueFn: func() Value {
					gen++
					return genValue(1)
				},
			},
			want:   "value1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &InMemoryCache{
				dataMutex: tt.fields.dataMutex,
				data:      tt.fields.data,
			}
			if got := cache.GetOrSet(tt.args.key, tt.args.valueFn); got != tt.want {
				t.Errorf("GetOrSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryCache_GetOrSet1(t *testing.T) {
	counter := 10
	var wg sync.WaitGroup
	var gen uint8
	c := NewInMemoryCache()

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			toStr := rand.Intn(counter) + 1
			key := strconv.Itoa(toStr)

			c.GetOrSet(key, func() Value {
				gen++
				return genValue(toStr)
			})
		}()
	}
	wg.Wait()

	if int(gen) != counter {
		t.Errorf("Генератор значений должен быть вызван %d раз, но был вызван %d", counter, gen)
	}

}

func TestInMemoryCache_GetOrSet2(t *testing.T) {
	c := NewInMemoryCache()
	var wg sync.WaitGroup

	toWrite := 10

	wg.Add(toWrite)
	// конкурентная запись новых значений
	for i := 0; i < toWrite; i++ {
		go func(counter int) {
			defer wg.Done()
			calls := 0 // счетчик вызовов генератора значений
			wait := genValue(counter)

			now := c.GetOrSet(strconv.Itoa(counter), func() Value {
				calls++
				return wait
			})

			if now != wait {
				t.Errorf("GetOrSet should return value, generated using the callback; expected: %s, got: %s", wait, now)
			}

			if calls != 1 {
				t.Errorf("value generator should be triggered only once, but have been called %d times", calls)
			}
		}(i)
	}
	wg.Wait()

	// Все происходит внутри одного теста, для того что бы не наполнять кэш данными заново
	// примерно точно таким же алгоритмом как выше, будем использовать уже готовые данные для следующих тестов
	// В какой то мере это не совсем правильно, но для того что бы не плодить одинаковый код написал так>

	wg.Add(toWrite)
	// Паралельная проверка на то, что все записанные значения присутствуют в кэше,
	// с правильным ключ-значением
	for i := 0; i < toWrite; i++ {
		go func(counter int) {
			defer wg.Done()

			generatorCalls := 0
			key := strconv.Itoa(counter)
			expected := genValue(counter)

			actual := c.GetOrSet(key, func() Value {
				generatorCalls++
				return "----------"
			})

			if actual != expected {
				t.Errorf("unexpected value received, expected is: %s, but got: %s", expected, actual)
			}
			if generatorCalls > 0 {
				t.Errorf("value generator should HAVE NOT been called, should return existing value from the cache")
			}
		}(i)
	}

	wg.Wait()
}
