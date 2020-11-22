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
	for i := 0; i < toWrite; i++ {
		go func(counter int) {
			defer wg.Done()
			calls := 0
			wait := genValue(counter)

			now := c.GetOrSet(strconv.Itoa(counter), func() Value {
				calls++
				return wait
			})

			if now != wait {
				t.Errorf("GetOrSet должен вернуть значение, ждем: %s, получили: %s", wait, now)
			}

			if calls != 1 {
				t.Errorf("генератор значений должен запускаться только один раз, но был вызван %d раз", calls)
			}
		}(i)
	}
	wg.Wait()
	wg.Add(toWrite)
	for i := 0; i < toWrite; i++ {
		go func(counter int) {
			defer wg.Done()

			gen := 0
			key := strconv.Itoa(counter)
			wait := genValue(counter)

			now:= c.GetOrSet(key, func() Value {
				gen++
				return "----------"
			})

			if now != wait {
				t.Errorf("получено неожиданное значение, ожидается: %s, но получено: %s", wait, now)
			}
			if gen > 0 {
				t.Errorf("генератор значений НЕ ДОЛЖЕН быть вызван, должен возвращать существующее значение из кеша")
			}
		}(i)
	}

	wg.Wait()
}
