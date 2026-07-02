package cache

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetGet(t *testing.T) {
	tests := []struct {
		name            string
		setKey          string
		setValue        any
		setTtl          time.Duration
		isErrorExpected bool
	}{
		{"simple set, get", "six seven", 67, 100 * time.Millisecond, false},
		{"negative ttl", "six seven", 67, -100 * time.Millisecond, true},
		{"null ttl", "six seven", 67, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CacheNew()

			err := c.Set(tt.setKey, tt.setValue, tt.setTtl)
			if tt.isErrorExpected {
				assert.Error(t, err)
			}
			actualValue, err := c.Get(tt.setKey)
			if tt.isErrorExpected {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.setValue, actualValue)
			}

			c.Stop()
		})
	}
}

func TestTtl(t *testing.T) {
	c := CacheNew()
	setKey := "six seven"
	setValue := 67
	setTtl := 100 * time.Millisecond

	// Через 50мс ключ доступен
	c.Set(setKey, setValue, setTtl)
	time.Sleep(50 * time.Millisecond)
	actualValue, err := c.Get(setKey)
	assert.NoError(t, err)
	assert.Equal(t, setValue, actualValue)

	// После истечения срока ключ не доступен
	time.Sleep(100 * time.Millisecond)
	actualValue, err = c.Get(setKey)
	assert.Error(t, err)
	assert.Equal(t, nil, actualValue)
	c.Stop()
}

func TestSetReplace(t *testing.T) {
	c := CacheNew()
	setKey := "key name"
	setValue1 := "old"
	setValue2 := "new"
	setTtl := 100 * time.Millisecond

	c.Set(setKey, setValue1, setTtl)
	c.Set(setKey, setValue2, setTtl)

	actualValue, err := c.Get(setKey)
	assert.NoError(t, err)
	assert.Equal(t, setValue2, actualValue)

	c.Stop()
}

func TestGetNull(t *testing.T) {
	c := CacheNew()
	setKey := "six seven"
	setValue := 67
	setTtl := 100 * time.Millisecond

	c.Set(setKey, setValue, setTtl)

	actualValue, err := c.Get("not a key")
	assert.Error(t, err)
	assert.Equal(t, nil, actualValue)

	c.Stop()
}

func TestDelete(t *testing.T) {
	c := CacheNew()
	setKey := "key"
	setValue1 := "new"
	setValue2 := "new"
	setTtl := 100 * time.Millisecond

	// При удалении несуществующего ключа не выбрасывается ошибки
	c.Delete("not a key")

	// После удаления ключа Get по тому же ключу возвращает nil, false
	c.Set(setKey, setValue1, setTtl)
	c.Delete(setKey)
	actualValue, err := c.Get(setKey)
	assert.Error(t, err)
	assert.Equal(t, nil, actualValue)

	// После удаления можно заново добвить элемент с тем же ключом
	c.Set(setKey, setValue2, setTtl)
	actualValue, err = c.Get(setKey)
	assert.NoError(t, err)
	assert.Equal(t, setValue2, actualValue)

	c.Stop()
}

func TestNewAndClear(t *testing.T) {
	c := CacheNew()

	// Инициализирует мапу `c.items != nil`
	assert.NotEqual(t, c.items, nil)

	// Мапа пустая `len(c.items) == 0`
	assert.Equal(t, len(c.items), 0)

	c.Clear()
	assert.NotEqual(t, c.items, nil)
	assert.Equal(t, len(c.items), 0)

	c.Stop()
}

func TestConcurency(t *testing.T) {
	c := CacheNew()
	testKeys := []string{
		"six seven",
		"six nine",
		"four twenty",
		"one one",
	}
	testValues := []any{
		"slaaaay",
		67,
		struct {
			name string
			age  int
		}{"Vasya", 99},
		"e05baf209aa9dad8d4ef58a09d3793672433dae2dfd6f9c1bca174cb462f1f31",
		nil,
	}
	testOperations := []string {
		"Set",
		"Get",
		"Delete",
	}

	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func(c *Cache) {
			time.Sleep(time.Millisecond * time.Duration(rand.IntN(100)))
			currentKey := testKeys[rand.IntN(len(testKeys))]
			currentValue := testValues[rand.IntN(len(testValues))]
			currentOperation := testOperations[rand.IntN(len(testOperations))]
			currentTtl := time.Millisecond * time.Duration(rand.IntN(100))

			switch currentOperation {
			case "Set":
				c.Set(currentKey, currentValue, currentTtl)
			case "Get":
				c.Get(currentKey)
			case "Delete":
				c.Delete(currentKey)
			}

			wg.Done()
		}(c)
	}

	wg.Wait()
	c.Stop()
}
