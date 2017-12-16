package lazysetup

import "fmt"

// Default setup object for project
var Default = New()

// Init all settings options, safe for second call (it will be ignored)
func Init() error {
	return Default.Init()
}

// OnInit without 'after' arguments appends setup function to setup chain functions ends. You can use 'after' args, to
// be sure that this setup function will be called only after setup steps which names listed in 'after' arguments.
func OnInit(setupFunc func() error, name string, after ...string) {
	Default.OnInit(setupFunc, name, after...)
}

// LazySetup configurator
type LazySetup struct {
	// onInits functions list
	onInits map[string]*initCall

	// alreadySetup flag
	alreadySetup bool
}

// New lazy setup instance constructor
func New() *LazySetup {
	return &LazySetup{
		onInits: make(map[string]*initCall),
	}
}

// OnInit without 'after' arguments appends setup function to setup chain functions ends. You can use 'after' args, to
// be sure that this setup function will be called only after setup steps which names listed in 'after' arguments.
func (s *LazySetup) OnInit(setupFunc func() error, name string, after ...string) {
	s.onInits[name] = &initCall{name: name, setupFunc: setupFunc, after: after}
}

// Init all settings options, safe for second call (it will be ignored)
func (s *LazySetup) Init() error {
	if s.alreadySetup {
		return nil
	}

	// lazy call all setup handlers, but resolve them in dependency order
	for key := range s.onInits {
		if err := s.resolve(map[string]struct{}{}, s.onInits[key]); err != nil {
			return err
		}
	}

	s.alreadySetup = true
	return nil
}

// resolve settings recurcive, path used to find cyclic dependecies
func (s *LazySetup) resolve(path map[string]struct{}, ic *initCall) error {
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
	return ic.setupFunc()
}

type initCall struct {
	name      string
	after     []string
	setupFunc func() error
	resolved  bool
}
