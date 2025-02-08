package testing

// Suite is an interface that runs a collection of tests.
type Suite interface {
	// Run executes the tests in the suite.
	Run(t *Tester)

	// Optional lifecycle hooks
	BeforeAll()
	AfterAll()
	BeforeEach()
	AfterEach()
}

// BaseSuite provides a default implementation of Suite.
type BaseSuite struct {
	runFn      func(t *Tester)
	beforeAll  func()
	afterAll   func()
	beforeEach func()
	afterEach  func()
}

// NewSuite creates a new Suite with the provided run function.
func NewSuite(run func(t *Tester)) *BaseSuite {
	return &BaseSuite{runFn: run}
}

// Run executes the tests in the suite.
func (s *BaseSuite) Run(t *Tester) {
	if s.beforeAll != nil {
		s.beforeAll()
	}
	if s.beforeEach != nil {
		s.beforeEach()
	}

	s.runFn(t)

	if s.afterEach != nil {
		s.afterEach()
	}
	if s.afterAll != nil {
		s.afterAll()
	}
}

// BeforeAll sets up the suite before any tests run.
func (s *BaseSuite) BeforeAll() {
	if s.beforeAll != nil {
		s.beforeAll()
	}
}

// AfterAll cleans up after all tests have run.
func (s *BaseSuite) AfterAll() {
	if s.afterAll != nil {
		s.afterAll()
	}
}

// BeforeEach runs before each test.
func (s *BaseSuite) BeforeEach() {
	if s.beforeEach != nil {
		s.beforeEach()
	}
}

// AfterEach runs after each test.
func (s *BaseSuite) AfterEach() {
	if s.afterEach != nil {
		s.afterEach()
	}
}

// SetBeforeAll sets the BeforeAll hook.
func (s *BaseSuite) SetBeforeAll(fn func()) {
	s.beforeAll = fn
}

// SetAfterAll sets the AfterAll hook.
func (s *BaseSuite) SetAfterAll(fn func()) {
	s.afterAll = fn
}

// SetBeforeEach sets the BeforeEach hook.
func (s *BaseSuite) SetBeforeEach(fn func()) {
	s.beforeEach = fn
}

// SetAfterEach sets the AfterEach hook.
func (s *BaseSuite) SetAfterEach(fn func()) {
	s.afterEach = fn
}
