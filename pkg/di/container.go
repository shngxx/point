package di

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// Container represents a simple DI container
type Container struct {
	mu         sync.RWMutex
	services   map[reflect.Type]any
	singletons map[reflect.Type]any
	providers  []providerInfo
}

// providerInfo stores information about a constructor
type providerInfo struct {
	constructor     reflect.Value
	constructorName string // name of the constructor function for better error messages
	paramTypes      []reflect.Type
	returnTypes     []reflect.Type
	returnsError    bool // indicates whether the constructor returns error as the last value
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services:   make(map[reflect.Type]any),
		singletons: make(map[reflect.Type]any),
		providers:  make([]providerInfo, 0),
	}
}

// Register registers a factory function for creating a service
func (c *Container) Register(serviceType reflect.Type, factory func() any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[serviceType] = factory
}

// RegisterSingleton registers a singleton service
func (c *Container) RegisterSingleton(serviceType reflect.Type, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons[serviceType] = instance
}

// resolve retrieves a service from the container (private method)
func (c *Container) resolve(serviceType reflect.Type) (any, error) {
	c.mu.RLock()

	// Check singleton
	if instance, ok := c.singletons[serviceType]; ok {
		c.mu.RUnlock()
		return instance, nil
	}

	// Check factory
	factory, ok := c.services[serviceType]
	c.mu.RUnlock()

	if !ok {
		// If an interface is requested, try to find an implementation
		if serviceType.Kind() == reflect.Interface {
			return c.resolveInterface(serviceType)
		}
		return nil, fmt.Errorf("service of type %v is not registered (use container.Supply() or container.Provide() to register it)", serviceType)
	}

	// Call factory
	factoryFunc := factory.(func() any)
	return factoryFunc(), nil
}

// resolveInterface attempts to find an interface implementation among registered types (private method)
func (c *Container) resolveInterface(interfaceType reflect.Type) (any, error) {
	c.mu.RLock()

	// Search among singletons
	for implType, instance := range c.singletons {
		if implType.Implements(interfaceType) {
			c.mu.RUnlock()
			return instance, nil
		}
	}

	// Search among registered services
	var factory func() any
	var found bool
	for implType, f := range c.services {
		if implType.Implements(interfaceType) {
			factory = f.(func() any)
			found = true
			break
		}
	}
	c.mu.RUnlock()

	if !found {
		return nil, fmt.Errorf("no implementation found for interface %v (register a type that implements this interface using container.Supply() or container.Provide())", interfaceType)
	}

	// Call factory outside of lock
	instance := factory()
	return instance, nil
}

// mustResolve retrieves a service from the container, panics on error (private method)
func (c *Container) mustResolve(serviceType reflect.Type) any {
	instance, err := c.resolve(serviceType)
	if err != nil {
		panic(err)
	}
	return instance
}

// MustResolve retrieves a service from the container by type (for internal use in container)
// Uses generics for type safety
func MustResolve[T any](container *Container) T {
	var zero T
	typ := reflect.TypeOf(&zero).Elem()
	instance := container.mustResolve(typ)
	return instance.(T)
}

// Supply registers ready values as singletons in the container.
// Unlike Provide, Supply accepts values directly, not constructors.
// Used for configuration, constants, and other ready values.
//
// Examples:
//   - container.Supply(DBConfig{Host: "localhost", Port: 5432})
//   - container.Supply(appConfig, redisConfig, serverConfig)
//
// Values are registered by their type and available for injection into constructors.
// Panics on errors.
func (c *Container) Supply(values ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, value := range values {
		if value == nil {
			panic(fmt.Errorf("Supply: value cannot be nil"))
		}

		valueType := reflect.TypeOf(value)

		// Check that it's not a function (use Provide for functions)
		if valueType.Kind() == reflect.Func {
			panic(fmt.Errorf("Supply: cannot accept functions, use Provide for constructors"))
		}

		// Check if this type is already registered
		if _, exists := c.singletons[valueType]; exists {
			panic(fmt.Errorf("Supply: value of type %v is already registered", valueType))
		}

		// Register value as singleton
		c.singletons[valueType] = value
	}
}

// Provide registers constructors for automatic dependency creation.
// Constructors can accept parameters (dependencies) and return one or more objects.
// Constructors can return error as the last value.
//
// Examples:
//   - func(*A, *B) *C
//   - func(*A, *B) (*C, error)
//   - func(*A) (*B, *C, error)
//
// Registration order doesn't matter. Constructors are called only if their types are needed.
// Results are cached (singleton within the container).
// Panics on errors.
func (c *Container) Provide(constructors ...any) {
	for _, constructor := range constructors {
		c.provideOne(constructor)
	}
}

// provideOne registers one constructor
func (c *Container) provideOne(constructor any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		panic(fmt.Errorf("Provide: constructor must be a function"))
	}

	// Analyze parameters (dependencies)
	numIn := constructorType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = constructorType.In(i)
	}

	// Analyze return values (provided services)
	numOut := constructorType.NumOut()
	if numOut == 0 {
		panic(fmt.Errorf("Provide: constructor must return at least one value"))
	}

	// Check if error is returned as the last value
	returnsError := false
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if numOut > 0 && constructorType.Out(numOut-1).Implements(errorInterface) {
		returnsError = true
	}

	// Collect return value types (excluding error)
	returnTypes := make([]reflect.Type, 0, numOut)
	for i := range numOut {
		returnType := constructorType.Out(i)
		// If this is the last element and it's error, skip it
		if i == numOut-1 && returnsError {
			continue
		}
		returnTypes = append(returnTypes, returnType)
	}

	if len(returnTypes) == 0 {
		panic(fmt.Errorf("Provide: constructor must return at least one non-error type"))
	}

	// Get constructor name for better error messages
	constructorName := getFunctionName(constructor)
	if constructorName == "" {
		// Fallback to return type if name cannot be determined
		if len(returnTypes) > 0 {
			constructorName = fmt.Sprintf("constructor returning %v", returnTypes[0])
		} else {
			constructorName = "constructor"
		}
	}

	// Save constructor information
	info := providerInfo{
		constructor:     reflect.ValueOf(constructor),
		constructorName: constructorName,
		paramTypes:      paramTypes,
		returnTypes:     returnTypes,
		returnsError:    returnsError,
	}
	c.providers = append(c.providers, info)

	// Register factories for each return type
	for idx, returnType := range returnTypes {
		// Create closure for each type (copy index and type to local variables)
		rt := returnType
		index := idx
		c.services[rt] = func() any {
			return c.invokeProviderForType(info, index, rt)
		}
	}
}

// invokeProviderForType invokes the constructor and returns a value of the required type
func (c *Container) invokeProviderForType(info providerInfo, returnIndex int, returnType reflect.Type) any {
	// Double-checked locking for thread-safe singleton creation
	c.mu.RLock()
	if instance, ok := c.singletons[returnType]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	// Lock for writing to create
	c.mu.Lock()

	// Check again (in case another thread already created it)
	if instance, ok := c.singletons[returnType]; ok {
		c.mu.Unlock()
		return instance
	}

	// Resolve dependencies (temporarily unlock mutex)
	args := make([]reflect.Value, len(info.paramTypes))
	for i, paramType := range info.paramTypes {
		// Temporarily unlock for dependency resolution
		c.mu.Unlock()
		instance, err := c.resolve(paramType)
		c.mu.Lock()
		if err != nil {
			c.mu.Unlock() // Unlock before panic
			paramName := fmt.Sprintf("parameter #%d", i+1)
			if len(info.paramTypes) == 1 {
				paramName = "parameter"
			}
			panic(fmt.Errorf("%s (%s) requires %s of type %v, but: %w",
				info.constructorName, returnType, paramName, paramType, err))
		}
		args[i] = reflect.ValueOf(instance)
	}

	// Unlock before calling constructor to avoid deadlock
	c.mu.Unlock()

	// Call constructor
	results := info.constructor.Call(args)

	// Check error if constructor returns it
	if info.returnsError {
		errorValue := results[len(results)-1]
		if !errorValue.IsNil() {
			// Constructor returned an error
			err := errorValue.Interface().(error)
			panic(fmt.Errorf("%s returned error: %w", info.constructorName, err))
		}
		// Remove error from results
		results = results[:len(results)-1]
	}

	// Lock again to save results
	c.mu.Lock()
	defer c.mu.Unlock()

	// Register all return values as singletons
	for i, result := range results {
		rt := info.returnTypes[i]
		// Check if someone created a singleton while we were calling the constructor
		if _, exists := c.singletons[rt]; !exists {
			c.singletons[rt] = result.Interface()
		}
	}

	// Return value of the required type
	if returnIndex < len(results) {
		return results[returnIndex].Interface()
	}
	return nil
}

// getFunctionName extracts the function name from a function value
func getFunctionName(fn any) string {
	if fn == nil {
		return ""
	}

	// Get function pointer
	pc := reflect.ValueOf(fn).Pointer()
	if pc == 0 {
		return ""
	}

	// Get function info
	funcInfo := runtime.FuncForPC(pc)
	if funcInfo == nil {
		return ""
	}

	// Get full function name (e.g., "github.com/user/pkg.NewService")
	fullName := funcInfo.Name()

	// Extract just the function name (e.g., "NewService" from "github.com/user/pkg.NewService")
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return fullName
}
