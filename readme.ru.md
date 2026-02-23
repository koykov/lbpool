# Leaky Buffer pool

Реализация `Pool` на базе шаблона [leaky buffer](https://golang.org/doc/effective_go.html#leaky_buffer). Вообще, этот
шаблон правильнее называть [leaky bucket](https://en.wikipedia.org/wiki/Leaky_bucket), но в терминах Go используется
именно такое название.

Это решение имеет как преимущества, так и недостатки перед нативной реализацией [sync.Pool](https://golang.org/src/sync/pool.go):
нативное решение более быстрое, но в противовес имеет серьёзный недостаток:
> Any item stored in the Pool may be removed automatically at any time without
notification. If the Pool holds the only reference when this happens, the
item might be deallocated.
> - https://golang.org/src/sync/pool.go

Если вам понадобится хранить в пуле что-то, что не может быть подчищено автоматически GC, то вы получите утечку памяти.
К таким объектам относятся, например [cbyte](https://github.com/koykov/cbyte) арены, которые требуют ручного освобождения.

Во всём остальном этот пул работает аналогично нативному.

## API

Метод `Get` пула аналогичен методу нативного пула и имеет сигнатуру:
```go
func Get() any
```

Метод `Put` способен принять только те объекты, которые реализуют интерфейс `Releaser`:
```go
type Releaser interface {
    Release()
}

func Put(x Releaser) bool
```
и также возвращает `bool` - протёк объект мимо ведра или нет.

Пул для использования требует инициализации через конструктор `New`:
```go
func New(size uint, releaseFactor float64, options ...Option) Pool
```
с аргументами:
* `size` - Максимальный размер пула. По его достижении все новые поступающие элементы будет протекать.
* `releaseFactor` - Доля поступающих элементов, которые принудительно протекут мимо пула. Это нужно для постепенного обновления элементов в пуле.
* `options` - Список опциональных функций, которые позволяют задать дополнительные параметры пула.

В данный момент доступны три опциональные функции:
* `WithNewFn` - Задаёт New функцию для конструирования объекта для возвращения вместо `nil` значения.
* `WithShards` - Количество шардов пула, нужно для снижения нагрузки на мьютекс внутреннего канала.
* `WithSampler` - Задаёт сэмплер для оценки возвращать элемент в пул или нет. Может быть полезен для обновления содержимого пула.
* `WithMetricsWriter` - Подключает к пулу компонент для записи метрик.

## Использование

```go
import (
    "github.com/koykov/lbpool"
    "github.com/koykov/lbpool/metrics/victoria"
)

const (
    poolSize      = 100 // max pool size
    releaseFactor = .05 // part of items that will leak independent of leaky logic
)

type item struct{...}

func (item) Release() {...}

var pool = lbpool.New(poolSize, releaseFactor, lbpool.WithNewFn(func() any { return &item{} }),
    lbpool.WithMetricsWriter(victoria.NewWriter("my_pool)))
    
itm := pool.Get().(*item)
...
pool.Put(itm)
```

## Производительность

Пул реализован поверх каналов Go и работает в несколько раз медленнее нативного пула:
```
BenchmarkPool-8                 	23870334	        50.07 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolParallel-8         	100000000	        11.62 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolNative-8           	100000000	        10.59 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolNativeParallel-8   	424538383	        2.920 ns/op	       0 B/op	       0 allocs/op
```
Учитывайте это при выборе инструментов под свои задачи.
