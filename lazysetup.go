package lazysetup

import "fmt"

// LazySettings configurator
type LazySettings struct {
	// onInits functions list
	onInits map[string]*callback

	// onClose functions list
	onClose map[string]*callback

	// alreadySetup flag
	alreadySetup bool

	// alreadyClosed flag
	alreadyClosed bool
}

// Default setup object for project
var Default = New()

// Init all settings options, safe for second call (it will be ignored)
func Init() error {
	return Default.Init()
}

// OnInit appends setupFunc function to list for lazy setup in default settings
func OnInit(setupFunc func() error, name string, after ...string) {
	Default.OnInit(setupFunc, name, after...)
}

// OnClose appends closeFunc function to list of defer closers in default settings
func OnClose(closeFunc func(), name string, after ...string) {
	Default.OnClose(closeFunc, name, after...)
}

// Close settings will call all registered close callbacks and don't return any error in default settings
func Close() {
	Default.Close()
}

// New instance constructor
func New() *LazySettings {
	return &LazySettings{
		onInits: make(map[string]*callback),
		onClose: make(map[string]*callback),
	}
}

// OnInit appends setupFunc function to list for lazy setup
func (s *LazySettings) OnInit(setupFunc func() error, name string, after ...string) {
	s.onInits[name] = &callback{name: name, hookFunc: setupFunc, after: after}
}

// OnClose appends closeFunc function to list of defer closers
func (s *LazySettings) OnClose(closeFunc func(), name string, after ...string) {
	s.onClose[name] = &callback{
		name: name,
		hookFunc: func() error {
			closeFunc()
			return nil
		},
		after: after,
	}
}

// Init all settings options, safe for second call (it will be ignored)
func (s *LazySettings) Init() error {
	// don't ignore errors, we can't continue if something not initialized
	return s.loopOverCallbacks(&s.alreadySetup, s.onInits)
}

// Close settings will call all registered close callbacks and don't return any error
func (s *LazySettings) Close() {
	// ignore errors, so last argument false
	s.loopOverCallbacks(&s.alreadyClosed, s.onClose)
}

// loopOverCallbacks functions and call them
func (s *LazySettings) loopOverCallbacks(flag *bool, cbList map[string]*callback) error {
	if *flag {
		return nil
	}

	// lazy call all callbacks, but resolve them in dependency order
	for key := range cbList {
		if err := s.resolve(map[string]struct{}{}, cbList[key]); err != nil {
			return err
		}
	}

	*flag = true
	return nil
}

// resolve settings recurcive, path used to find cyclic dependecies
func (s *LazySettings) resolve(path map[string]struct{}, ic *callback) error {
	if ic.resolved {
		return nil
	}
	path[ic.name] = struct{}{}

	// recurcive resolve dependencies
	for _, name := range ic.after {
		dependency := s.onInits[name]
		if !dependency.resolved {
			if _, ok := path[name]; ok {
				return fmt.Errorf("'%v' has cyclic dependency '%v'", ic.name, name)
			}
			if err := s.resolve(path, dependency); err != nil {
				return err
			}
		}
	}
	ic.resolved = true
	return ic.hookFunc()
}

type callback struct {
	name     string
	after    []string
	hookFunc func() error
	resolved bool
}
