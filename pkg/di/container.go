package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container представляет простой DI контейнер
type Container struct {
	mu         sync.RWMutex
	services   map[reflect.Type]interface{}
	singletons map[reflect.Type]interface{}
	providers  []providerInfo
}

// providerInfo хранит информацию о конструкторе
type providerInfo struct {
	constructor reflect.Value
	paramTypes  []reflect.Type
	returnTypes []reflect.Type
}

// NewContainer создает новый DI контейнер
func NewContainer() *Container {
	return &Container{
		services:   make(map[reflect.Type]interface{}),
		singletons: make(map[reflect.Type]interface{}),
		providers:  make([]providerInfo, 0),
	}
}

// Register регистрирует фабричную функцию для создания сервиса
func (c *Container) Register(serviceType reflect.Type, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[serviceType] = factory
}

// RegisterSingleton регистрирует singleton сервис
func (c *Container) RegisterSingleton(serviceType reflect.Type, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons[serviceType] = instance
}

// resolve извлекает сервис из контейнера (приватный метод)
func (c *Container) resolve(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()

	// Проверяем singleton
	if instance, ok := c.singletons[serviceType]; ok {
		c.mu.RUnlock()
		return instance, nil
	}

	// Проверяем фабрику
	factory, ok := c.services[serviceType]
	c.mu.RUnlock()

	if !ok {
		// Если запрашивается интерфейс, пытаемся найти реализацию
		if serviceType.Kind() == reflect.Interface {
			return c.resolveInterface(serviceType)
		}
		return nil, fmt.Errorf("сервис типа %v не зарегистрирован", serviceType)
	}

	// Вызываем фабрику
	factoryFunc := factory.(func() interface{})
	return factoryFunc(), nil
}

// resolveInterface пытается найти реализацию интерфейса среди зарегистрированных типов (приватный метод)
func (c *Container) resolveInterface(interfaceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ищем среди singletons
	for implType, instance := range c.singletons {
		if implType.Implements(interfaceType) {
			return instance, nil
		}
	}

	// Ищем среди зарегистрированных сервисов
	for implType, factory := range c.services {
		if implType.Implements(interfaceType) {
			c.mu.RUnlock()
			factoryFunc := factory.(func() interface{})
			instance := factoryFunc()
			c.mu.RLock()
			return instance, nil
		}
	}

	return nil, fmt.Errorf("не найдена реализация интерфейса %v", interfaceType)
}

// mustResolve извлекает сервис из контейнера, паникует при ошибке (приватный метод)
func (c *Container) mustResolve(serviceType reflect.Type) interface{} {
	instance, err := c.resolve(serviceType)
	if err != nil {
		panic(err)
	}
	return instance
}

// MustResolveType извлекает сервис из контейнера по типу (для внутреннего использования в container)
// Использует дженерики для типобезопасности
func MustResolveType[T any](container *Container) T {
	var zero T
	typ := reflect.TypeOf(&zero).Elem()
	instance := container.mustResolve(typ)
	return instance.(T)
}

// Provide регистрирует конструктор (функцию) для автоматического создания зависимостей
// Конструктор может принимать параметры (зависимости) и возвращать значения (предоставляемые сервисы)
// Пример: container.Provide(func(service MyService) *Handler { return NewHandler(service) })
func (c *Container) Provide(constructor interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("Provide: конструктор должен быть функцией")
	}

	// Анализируем параметры (зависимости)
	numIn := constructorType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = constructorType.In(i)
	}

	// Анализируем возвращаемые значения (предоставляемые сервисы)
	numOut := constructorType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("Provide: конструктор должен возвращать хотя бы одно значение")
	}
	returnTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		returnTypes[i] = constructorType.Out(i)
	}

	// Сохраняем информацию о конструкторе
	info := providerInfo{
		constructor: reflect.ValueOf(constructor),
		paramTypes:  paramTypes,
		returnTypes: returnTypes,
	}
	c.providers = append(c.providers, info)

	// Регистрируем фабрики для каждого возвращаемого типа
	for idx, returnType := range returnTypes {
		// Создаем замыкание для каждого типа (копируем индекс и тип в локальные переменные)
		rt := returnType
		index := idx
		c.services[rt] = func() interface{} {
			return c.invokeProviderForType(info, index, rt)
		}
	}

	return nil
}

// invokeProviderForType вызывает конструктор и возвращает значение нужного типа
func (c *Container) invokeProviderForType(info providerInfo, returnIndex int, returnType reflect.Type) interface{} {
	// Double-checked locking для thread-safe создания singleton
	c.mu.RLock()
	if instance, ok := c.singletons[returnType]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	// Блокируем на запись для создания
	c.mu.Lock()
	defer c.mu.Unlock()

	// Проверяем еще раз (на случай, если другой поток уже создал)
	if instance, ok := c.singletons[returnType]; ok {
		return instance
	}

	// Резолвим зависимости (без блокировки, т.к. мы уже в критической секции)
	args := make([]reflect.Value, len(info.paramTypes))
	for i, paramType := range info.paramTypes {
		// Временно разблокируем для резолва зависимостей
		c.mu.Unlock()
		instance, err := c.resolve(paramType)
		c.mu.Lock()
		if err != nil {
			panic(fmt.Errorf("не удалось резолвить зависимость типа %v: %w", paramType, err))
		}
		args[i] = reflect.ValueOf(instance)
	}

	// Вызываем конструктор
	results := info.constructor.Call(args)

	// Регистрируем все возвращаемые значения как singletons
	for i, result := range results {
		rt := info.returnTypes[i]
		c.singletons[rt] = result.Interface()
	}

	// Возвращаем значение нужного типа
	if returnIndex < len(results) {
		return results[returnIndex].Interface()
	}
	return nil
}
